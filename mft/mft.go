package mft

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/t9t/gomft/binutil"
	"github.com/t9t/gomft/fragment"
	"github.com/t9t/gomft/utf16"
)

const (
	AttributeTypeStandardInformation AttributeType = 0x10 // always resident
	AttributeTypeAttributeList       AttributeType = 0x20 // mixed residency
	AttributeTypeFileName            AttributeType = 0x30 // always resident
	AttributeTypeObjectId            AttributeType = 0x40 // always resident
	AttributeTypeSecurityDescriptor  AttributeType = 0x50 // always resident?
	AttributeTypeVolumeName          AttributeType = 0x60 // always resident?
	AttributeTypeVolumeInformation   AttributeType = 0x70 // never resident?
	AttributeTypeData                AttributeType = 0x80 // mixed residency
	AttributeTypeIndexRoot           AttributeType = 0x90 // always resident
	AttributeTypeIndexAllocation     AttributeType = 0xa0 // never resident?
	AttributeTypeBitmap              AttributeType = 0xb0 // nearly always resident?
	AttributeTypeReparsePoint        AttributeType = 0xc0 // always resident?
	AttributeTypeEAInformation       AttributeType = 0xd0 // always resident
	AttributeTypeEA                  AttributeType = 0xe0 // nearly always resident?
	AttributeTypePropertySet         AttributeType = 0xf0
	AttributeTypeLoggedUtilityStream AttributeType = 0x100 // always resident
	AttributeTypeTerminator          AttributeType = 0xFFFFFFFF
)

var (
	fileSignature = []byte{0x46, 0x49, 0x4c, 0x45}
)

type Record struct {
	Header     RecordHeader
	Attributes []Attribute
}

func ParseRecord(b []byte) (Record, error) {
	b = binutil.Duplicate(b)
	header, err := ParseRecordHeader(b)
	if err != nil {
		return Record{}, err
	}
	f := header.FirstAttributeOffset
	if f < 0 || f >= len(b) {
		return Record{}, fmt.Errorf("invalid first attribute offset %d (data length: %d)", f, len(b))
	}

	b, err = applyFixUp(b, header.UpdateSequenceOffset, header.UpdateSequenceSize)
	if err != nil {
		return Record{}, fmt.Errorf("unable to apply fixup: %w", err)
	}

	attributes, err := ParseAttributes(b[f:])
	if err != nil {
		return Record{}, err
	}
	return Record{Header: header, Attributes: attributes}, nil
}

type RecordHeader struct {
	Signature             []byte
	UpdateSequenceOffset  int
	UpdateSequenceSize    int
	LogFileSequenceNumber uint64
	SequenceNumber        int
	HardLinkCount         int
	FirstAttributeOffset  int
	Flags                 RecordFlag
	ActualSize            uint32
	AllocatedSize         uint32
	BaseRecordReference   FileReference
	NextAttributeId       int
	RecordNumber          uint32
}

type FileReference struct {
	RecordNumber   uint64
	SequenceNumber uint16
}

func ParseFileReference(b []byte) (FileReference, error) {
	if len(b) != 8 {
		return FileReference{}, fmt.Errorf("expected 8 bytes but got %d", len(b))
	}

	return FileReference{
		RecordNumber:   binary.LittleEndian.Uint64(padTo(b[:6], 8)),
		SequenceNumber: binary.LittleEndian.Uint16(b[6:]),
	}, nil
}

type RecordFlag uint16

const (
	FlagInUse       RecordFlag = 0x0001
	FlagIsDirectory RecordFlag = 0x0002
	FlagInExtend    RecordFlag = 0x0004
	FlagIsIndex     RecordFlag = 0x0008
)

func ParseRecordHeader(b []byte) (RecordHeader, error) {
	if len(b) < 42 {
		return RecordHeader{}, fmt.Errorf("record header data length should be at least 42 but is %d", len(b))
	}
	sig := b[:4]
	if bytes.Compare(sig, fileSignature) != 0 {
		return RecordHeader{}, fmt.Errorf("unknown record signature: %# x", sig)
	}
	r := binutil.NewLittleEndianReader(b)
	baseRecordRef, err := ParseFileReference(r.Read(0x20, 8))
	if err != nil {
		return RecordHeader{}, fmt.Errorf("unable to parse base record reference: %v", err)
	}
	return RecordHeader{
		Signature:             binutil.Duplicate(sig),
		UpdateSequenceOffset:  int(r.Uint16(0x04)),
		UpdateSequenceSize:    int(r.Uint16(0x06)),
		LogFileSequenceNumber: r.Uint64(0x08),
		SequenceNumber:        int(r.Uint16(0x10)),
		HardLinkCount:         int(r.Uint16(0x12)),
		FirstAttributeOffset:  int(r.Uint16(0x14)),
		Flags:                 RecordFlag(r.Uint16(0x16)),
		ActualSize:            r.Uint32(0x18),
		AllocatedSize:         r.Uint32(0x1C),
		BaseRecordReference:   baseRecordRef,
		NextAttributeId:       int(r.Uint16(0x28)),
		RecordNumber:          r.Uint32(0x2C),
	}, nil
}

func applyFixUp(b []byte, offset int, length int) ([]byte, error) {
	r := binutil.NewLittleEndianReader(b)

	updateSequence := r.Read(offset, length*2) // length is in pairs, not bytes
	updateSequenceNumber := updateSequence[:2]
	updateSequenceArray := updateSequence[2:]

	sectorCount := len(updateSequenceArray) / 2
	sectorSize := len(b) / sectorCount

	for i := 1; i <= sectorCount; i++ {
		offset := sectorSize*i - 2
		if bytes.Compare(updateSequenceNumber, b[offset:offset+2]) != 0 {
			return nil, fmt.Errorf("update sequence mismatch at pos %d", offset)
		}
	}

	for i := 0; i < sectorCount; i++ {
		offset := sectorSize*(i+1) - 2
		num := i * 2
		copy(b[offset:offset+2], updateSequenceArray[num:num+2])
	}

	return b, nil
}

type Attribute struct {
	Type        AttributeType
	Resident    bool
	Name        string
	Flags       AttributeFlags
	AttributeId int
	Data        []byte
}

type AttributeType uint32
type AttributeFlags []byte

func ParseAttributes(b []byte) ([]Attribute, error) {
	if len(b) == 0 {
		return []Attribute{}, nil
	}
	attributes := make([]Attribute, 0)
	for len(b) > 0 {
		if len(b) < 4 {
			return nil, fmt.Errorf("attribute header data should be at least 4 bytes but is %d", len(b))
		}

		r := binutil.NewLittleEndianReader(b)
		attrType := r.Uint32(0)
		if attrType == uint32(AttributeTypeTerminator) {
			break
		}

		if len(b) < 8 {
			return nil, fmt.Errorf("cannot read attribute header record length, data should be at least 8 bytes but is %d", len(b))
		}

		recordLength := int(r.Uint32(0x04))
		if recordLength <= 0 {
			return nil, fmt.Errorf("cannot handle attribute with zero or negative record length %d", recordLength)
		}

		if recordLength > len(b) {
			return nil, fmt.Errorf("attribute record length %d exceeds data length %d", recordLength, len(b))
		}

		recordData := r.Read(0, recordLength)
		attribute, err := ParseAttribute(recordData)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attribute)
		b = r.ReadFrom(recordLength)
	}
	return attributes, nil
}

func ParseAttribute(b []byte) (Attribute, error) {
	if len(b) < 22 {
		return Attribute{}, fmt.Errorf("attribute data should be at least 22 bytes but is %d", len(b))
	}

	r := binutil.NewLittleEndianReader(b)

	nameLength := r.Byte(0x09)
	nameOffset := r.Uint16(0x0A)

	name := ""
	if nameLength != 0 {
		nameBytes := r.Read(int(nameOffset), int(nameLength)*2)
		decoded, err := utf16.DecodeString(nameBytes, binary.LittleEndian)
		if err != nil {
			return Attribute{}, fmt.Errorf("unable to parse attribute name: %w", err)
		}
		name = decoded
	}

	resident := r.Byte(0x08) == 0x00
	var attributeData []byte
	if resident {
		dataOffset := int(r.Uint16(0x14))
		dataLength := int(r.Uint32(0x10))
		expectedDataLength := dataOffset + dataLength

		if len(b) < expectedDataLength {
			return Attribute{}, fmt.Errorf("expected attribute data length to be at least %d but is %d", expectedDataLength, len(b))
		}

		attributeData = r.Read(dataOffset, dataLength)
	} else {
		dataOffset := int(r.Uint16(0x20))
		if len(b) < dataOffset {
			return Attribute{}, fmt.Errorf("expected attribute data length to be at least %d but is %d", dataOffset, len(b))
		}
		attributeData = r.ReadFrom(int(dataOffset))
	}

	return Attribute{
		Type:        AttributeType(r.Uint32(0)),
		Resident:    resident,
		Name:        name,
		Flags:       AttributeFlags(binutil.Duplicate(r.Read(0x0C, 2))),
		AttributeId: int(r.Uint16(0x0E)),
		Data:        binutil.Duplicate(attributeData),
	}, nil
}

func (a *Attribute) ParseDataAsStandardInformation() (StandardInformation, error) {
	if a.Type != AttributeTypeStandardInformation {
		return StandardInformation{}, fmt.Errorf("attribute type %#x is not $STANDARD_INFORMATION (%#x)", a.Type, AttributeTypeStandardInformation)
	}
	if !a.Resident {
		return StandardInformation{}, fmt.Errorf("cannot deal with non-resident $STANDARD_INFORMATION attribute")
	}

	return ParseStandardInformation(a.Data)
}

func (a *Attribute) ParseDataAsFileName() (FileName, error) {
	if a.Type != AttributeTypeFileName {
		return FileName{}, fmt.Errorf("attribute type %#x is not $FILE_NAME (%#x)", a.Type, AttributeTypeFileName)
	}
	if !a.Resident {
		return FileName{}, fmt.Errorf("cannot deal with non-resident $FILE_NAME attribute")
	}

	return ParseFileName(a.Data)
}

type DataRun struct {
	OffsetCluster    int64
	LengthInClusters uint64
}

func ParseDataRuns(b []byte) ([]DataRun, error) {
	if len(b) == 0 {
		return []DataRun{}, nil
	}

	runs := make([]DataRun, 0)
	for len(b) > 0 {
		r := binutil.NewLittleEndianReader(b)
		header := r.Byte(0)
		if header == 0 {
			break
		}

		lengthLength := int(header &^ 0xF0)
		offsetLength := int(header >> 4)

		dataRunDataLength := offsetLength + lengthLength

		headerAndDataLength := dataRunDataLength + 1
		if len(b) < headerAndDataLength {
			return nil, fmt.Errorf("expected at least %d bytes of datarun data but is %d", headerAndDataLength, len(b))
		}

		dataRunData := r.Reader(1, dataRunDataLength)

		lengthBytes := dataRunData.Read(0, lengthLength)
		dataLength := binary.LittleEndian.Uint64(padTo(lengthBytes, 8))

		offsetBytes := dataRunData.Read(lengthLength, offsetLength)
		dataOffset := int64(binary.LittleEndian.Uint64(padTo(offsetBytes, 8)))

		runs = append(runs, DataRun{OffsetCluster: dataOffset, LengthInClusters: dataLength})

		b = r.ReadFrom(headerAndDataLength)
	}

	return runs, nil
}

func DataRunsToFragments(runs []DataRun, bytesPerCluster int) []fragment.Fragment {
	frags := make([]fragment.Fragment, len(runs))
	previousOffsetCluster := int64(0)
	for i, run := range runs {
		exactClusterOffset := previousOffsetCluster + run.OffsetCluster
		frags[i] = fragment.Fragment{
			Offset: exactClusterOffset * int64(bytesPerCluster),
			Length: int(run.LengthInClusters) * bytesPerCluster,
		}
		previousOffsetCluster = exactClusterOffset
	}
	return frags
}

func padTo(data []byte, length int) []byte {
	dl := len(data)
	if dl > length {
		return data
	}
	if dl == length {
		return data
	}
	result := make([]byte, length)
	copy(result, data)
	if data[dl-1]&0b10000000 == 0b10000000 {
		for i := dl; i < length; i++ {
			result[i] = 0xFF
		}
	}
	return result
}

func (at AttributeType) Name() string {
	switch at {
	case AttributeTypeStandardInformation:
		return "$STANDARD_INFORMATION"
	case AttributeTypeAttributeList:
		return "$ATTRIBUTE_LIST"
	case AttributeTypeFileName:
		return "$FILE_NAME"
	case AttributeTypeObjectId:
		return "$OBJECT_ID"
	case AttributeTypeSecurityDescriptor:
		return "$SECURITY_DESCRIPTOR"
	case AttributeTypeVolumeName:
		return "$VOLUME_NAME"
	case AttributeTypeVolumeInformation:
		return "$VOLUME_INFORMATION"
	case AttributeTypeData:
		return "$DATA"
	case AttributeTypeIndexRoot:
		return "$INDEX_ROOT"
	case AttributeTypeIndexAllocation:
		return "$INDEX_ALLOCATION"
	case AttributeTypeBitmap:
		return "$BITMAP"
	case AttributeTypeReparsePoint:
		return "$REPARSE_POINT"
	case AttributeTypeEAInformation:
		return "$EA_INFORMATION"
	case AttributeTypeEA:
		return "$EA"
	case AttributeTypePropertySet:
		return "$PROPERTY_SET"
	case AttributeTypeLoggedUtilityStream:
		return "$LOGGED_UTILITY_STREAM"
	}
	return "unknown"
}
