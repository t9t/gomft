package mft

import (
	"github.com/t9t/gomft/binutil"
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
	r := binutil.BinReader(b)
	return RecordHeader{
		Signature:             r.Read(0, 4),
		UpdateSequenceOffset:  int(r.Uint16(0x04)),
		UpdateSequenceSize:    int(r.Uint16(0x06)),
		LogFileSequenceNumber: r.Uint64(0x08),
		RecordUsageNumber:     int(r.Uint16(0x10)),
		HardLinkCount:         int(r.Uint16(0x12)),
		FirstAttributeOffset:  int(r.Uint16(0x14)),
		Flags:                 r.Read(0x16, 2),
		ActualSize:            r.Uint32(0x18),
		AllocatedSize:         r.Uint32(0x1C),
		BaseRecordReference:   r.Read(0x20, 8),
		NextAttributeId:       int(r.Uint16(0x28)),
	}
}
