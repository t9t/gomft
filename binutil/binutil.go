package binutil

import "encoding/binary"

type BinReader []byte
var endian = binary.LittleEndian

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
