package mft

import (
	"encoding/binary"
	"fmt"

	"github.com/t9t/gomft/binutil"
	"github.com/t9t/gomft/utf16"
)

const (
	ATTRIBUTE_TYPE_STANDARD_INFORMATION  AttributeType = 0x10
	ATTRIBUTE_TYPE_ATTRIBUTE_LIST        AttributeType = 0x20
	ATTRIBUTE_TYPE_FILE_NAME             AttributeType = 0x30
	ATTRIBUTE_TYPE_OBJECT_ID             AttributeType = 0x40
	ATTRIBUTE_TYPE_VOLUME_NAME           AttributeType = 0x60
	ATTRIBUTE_TYPE_VOLUME_INFORMATION    AttributeType = 0x70
	ATTRIBUTE_TYPE_DATA                  AttributeType = 0x80
	ATTRIBUTE_TYPE_INDEX_ROOT            AttributeType = 0x90
	ATTRIBUTE_TYPE_INDEX_ALLOCATION      AttributeType = 0xA0
	ATTRIBUTE_TYPE_BITMAP                AttributeType = 0xB0
	ATTRIBUTE_TYPE_REPARSE_POINT         AttributeType = 0xC0
	ATTRIBUTE_TYPE_EA_INFORMATION        AttributeType = 0xD0
	ATTRIBUTE_TYPE_EA                    AttributeType = 0xE0
	ATTRIBUTE_TYPE_PROPERTY_SET          AttributeType = 0xF0
	ATTRIBUTE_TYPE_LOGGED_UTILITY_STREAM AttributeType = 0x100
	ATTRIBUTE_TYPE_TERMINATOR            AttributeType = 0xFFFFFFFF
)

type Record struct {
	Header     RecordHeader
	Attributes []Attribute
}

func ParseRecord(b []byte) (Record, error) {
	header, err := ParseRecordHeader(b)
	if err != nil {
		return Record{}, err
	}
	f := header.FirstAttributeOffset
	if f < 0 || f >= len(b) {
		return Record{}, fmt.Errorf("invalid first attribute offset %d (data length: %d)", f, len(b))
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
	RecordUsageNumber     int
	HardLinkCount         int
	FirstAttributeOffset  int
	Flags                 Flags
	ActualSize            uint32
	AllocatedSize         uint32
	BaseRecordReference   FileReference
	NextAttributeId       int
}

type FileReference []byte
type Flags []byte

func ParseRecordHeader(b []byte) (RecordHeader, error) {
	if len(b) < 42 {
		return RecordHeader{}, fmt.Errorf("record header data length should be at least 42 but is %d", len(b))
	}
	r := binutil.NewLittleEndianReader(b)
	return RecordHeader{
		Signature:             binutil.Duplicate(r.Read(0, 4)),
		UpdateSequenceOffset:  int(r.Uint16(0x04)),
		UpdateSequenceSize:    int(r.Uint16(0x06)),
		LogFileSequenceNumber: r.Uint64(0x08),
		RecordUsageNumber:     int(r.Uint16(0x10)),
		HardLinkCount:         int(r.Uint16(0x12)),
		FirstAttributeOffset:  int(r.Uint16(0x14)),
		Flags:                 binutil.Duplicate(r.Read(0x16, 2)),
		ActualSize:            r.Uint32(0x18),
		AllocatedSize:         r.Uint32(0x1C),
		BaseRecordReference:   binutil.Duplicate(r.Read(0x20, 8)),
		NextAttributeId:       int(r.Uint16(0x28)),
	}, nil
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
	attributes := make([]Attribute, 0)
	for len(b) > 0 {
		r := binutil.NewLittleEndianReader(b)
		attrType := r.Uint32(0)
		if attrType == uint32(ATTRIBUTE_TYPE_TERMINATOR) {
			break
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
	r := binutil.NewLittleEndianReader(b)

	nameLength := r.Byte(0x09)
	nameOffset := r.Uint16(0x0A)

	nameData := r.Read(int(nameOffset), int(nameLength))
	nameStr, err := utf16.DecodeString(nameData, r.ByteOrder())
	if err != nil {
		return Attribute{}, fmt.Errorf("unable to decode UTF16 attribute name: %w", err)
	}

	resident := r.Byte(0x08) == 0x00
	dataOffset := 0x40 //non-resident
	if resident {
		dataOffset = 0x18
	}

	return Attribute{
		Type:        AttributeType(r.Uint32(0)),
		Resident:    resident,
		Name:        nameStr,
		Flags:       AttributeFlags(binutil.Duplicate(r.Read(0x0C, 2))),
		AttributeId: int(r.Uint16(0x0E)),
		Data:        binutil.Duplicate(r.ReadFrom(dataOffset)),
	}, nil
}

type DataRun struct {
	OffsetCluster    int64
	LengthInClusters uint64
}

func ParseDataRuns(b []byte) ([]DataRun, error) {
	if len(b) == 0 {
		return []DataRun{}, nil
	}

	r := binutil.NewLittleEndianReader(b)
	runs := make([]DataRun, 0)
	for r.Length() > 0 {
		header := r.Byte(0)
		if header == 0 {
			break
		}

		lengthLength := int(header &^ 0xF0)
		offsetLength := int(header >> 4)

		dataRunDataLength := offsetLength + lengthLength
		dataRunData := r.Reader(1, dataRunDataLength)

		lengthBytes := dataRunData.Read(0, lengthLength)
		dataLength := binary.LittleEndian.Uint64(padTo(lengthBytes, 8))

		offsetBytes := dataRunData.Read(lengthLength, offsetLength)
		dataOffset := int64(binary.LittleEndian.Uint64(padTo(offsetBytes, 8)))

		runs = append(runs, DataRun{OffsetCluster: dataOffset, LengthInClusters: dataLength})
		r = r.ReaderFrom(dataRunDataLength + 1)
	}

	return runs, nil
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
	if data[dl-1] & 0b10000000 == 0b10000000 {
		for i := dl; i < length; i++ {
			result[i] = 0xFF
		}
	}
	return result
}
