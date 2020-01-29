package bootsect

import "encoding/binary"

type BootSector struct {
	OemId                              string
	BytesPerSector                     int
	SectorsPerCluster                  int
	MediaDescriptor                    byte
	SectorsPerTrack                    int
	NumberofHeads                      int
	HiddenSectors                      int
	TotalSectors                       uint64
	MftClusterNumber                   uint64
	MftMirrorClusterNumber             uint64
	BytesOrClusterPerFileSecordSegment byte
	BytersOrClusterPerIndexBuffer      byte
	VolumeSerialNumber                 []byte
}

var endian = binary.LittleEndian

func Parse(data []byte) BootSector {
	r := BinReader(data)
	return BootSector{
		OemId:                              string(r.Read(0x03, 8)),
		BytesPerSector:                     int(r.Uint16(0x0B)),
		SectorsPerCluster:                  int(r[0x0D]),
		MediaDescriptor:                    r[0x15],
		SectorsPerTrack:                    int(r.Uint16(0x18)),
		NumberofHeads:                      int(r.Uint16(0x1A)),
		HiddenSectors:                      int(r.Uint16(0x1C)),
		TotalSectors:                       r.Uint64(0x28),
		MftClusterNumber:                   r.Uint64(0x30),
		MftMirrorClusterNumber:             r.Uint64(0x38),
		BytesOrClusterPerFileSecordSegment: r[0x40],
		BytersOrClusterPerIndexBuffer:      r[0x44],
		VolumeSerialNumber:                 dupe(r.Read(0x48, 8)),
	}
}

func dupe(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

type BinReader []byte

func (b BinReader) Read(offset int, length int) []byte {
	return b[offset : offset+length]
}

func (b BinReader) Uint16(offset int) uint16 {
	return endian.Uint16(b.Read(offset, 2))
}

func (b BinReader) Uint32(offset int) uint32 {
	return endian.Uint32(b.Read(offset, 4))
}

func (b BinReader) Uint64(offset int) uint64 {
	return endian.Uint64(b.Read(offset, 8))
}
