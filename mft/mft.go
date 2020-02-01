package mft

import (
	"fmt"

	"github.com/t9t/gomft/binutil"
	"github.com/t9t/gomft/utf16"
)

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

func ParseRecordHeader(b []byte) RecordHeader {
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
	}
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
		Flags:       AttributeFlags(r.Read(0x0C, 2)),
		AttributeId: int(r.Uint16(0x0E)),
		Data:        binutil.Duplicate(r.ReadFrom(dataOffset)),
	}, nil
}
