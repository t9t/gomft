package bootsect

import (
	"fmt"
	"github.com/t9t/gomft/binutil"
)

type BootSector struct {
	OemId                  string
	BytesPerSector         int
	SectorsPerCluster      int
	MediaDescriptor        byte
	SectorsPerTrack        int
	NumberofHeads          int
	HiddenSectors          int
	TotalSectors           uint64
	MftClusterNumber       uint64
	MftMirrorClusterNumber uint64
	FileRecordSegmentSize  BytesOrClusters
	IndexBufferSize        BytesOrClusters
	VolumeSerialNumber     []byte
}

type BytesOrClusters struct {
	IsBytes bool
	Value   int
}

func Parse(data []byte) (BootSector, error) {
	if len(data) < 80 {
		return BootSector{}, fmt.Errorf("boot sector data should be at least 80 bytes but is %d", len(data))
	}
	r := binutil.NewLittleEndianReader(data)
	return BootSector{
		OemId:                  string(r.Read(0x03, 8)),
		BytesPerSector:         int(r.Uint16(0x0B)),
		SectorsPerCluster:      int(r.Byte(0x0D)),
		MediaDescriptor:        r.Byte(0x15),
		SectorsPerTrack:        int(r.Uint16(0x18)),
		NumberofHeads:          int(r.Uint16(0x1A)),
		HiddenSectors:          int(r.Uint16(0x1C)),
		TotalSectors:           r.Uint64(0x28),
		MftClusterNumber:       r.Uint64(0x30),
		MftMirrorClusterNumber: r.Uint64(0x38),
		FileRecordSegmentSize:  parseBytesOrClusters(r.Byte(0x40)),
		IndexBufferSize:        parseBytesOrClusters(r.Byte(0x44)),
		VolumeSerialNumber:     binutil.Duplicate(r.Read(0x48, 8)),
	}, nil
}

func parseBytesOrClusters(b byte) BytesOrClusters {
	i := int(int8(b))
	if i < 0 {
		value := 1 << -i
		return BytesOrClusters{IsBytes: true, Value: value}
	}
	return BytesOrClusters{IsBytes: false, Value: int(b)}
}

func (boc *BytesOrClusters) ToBytes(bytesPerCluster int) int {
	if boc.IsBytes {
		return boc.Value
	}
	return boc.Value * bytesPerCluster
}
