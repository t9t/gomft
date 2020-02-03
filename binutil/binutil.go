package binutil

import "encoding/binary"

func Duplicate(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

type BinReader struct {
	data []byte
	bo   binary.ByteOrder
}

func NewBinReader(data []byte, bo binary.ByteOrder) *BinReader {
	return &BinReader{data: data, bo: bo}
}

func NewLittleEndianReader(data []byte) *BinReader {
	return NewBinReader(data, binary.LittleEndian)
}

func NewBigEndianReader(data []byte, bo binary.ByteOrder) *BinReader {
	return NewBinReader(data, binary.BigEndian)
}

func (r *BinReader) Data() []byte {
	return r.data
}

func (r *BinReader) ByteOrder() binary.ByteOrder {
	return r.bo
}

func (r *BinReader) Length() int {
	return len(r.data)
}

func (r *BinReader) Read(offset int, length int) []byte {
	return r.data[offset : offset+length]
}

func (r *BinReader) Reader(offset int, length int) *BinReader {
	return &BinReader{data: r.data[offset : offset+length], bo: r.bo}
}

func (r *BinReader) Byte(offset int) byte {
	return r.Read(offset, 1)[0]
}

func (r *BinReader) ReadFrom(offset int) []byte {
	return r.data[offset:]
}

func (r *BinReader) ReaderFrom(offset int) *BinReader {
	return &BinReader{data: r.data[offset:], bo: r.bo}
}

func (r *BinReader) Uint16(offset int) uint16 {
	return r.bo.Uint16(r.Read(offset, 2))
}

func (r *BinReader) Uint32(offset int) uint32 {
	return r.bo.Uint32(r.Read(offset, 4))
}

func (r *BinReader) Uint64(offset int) uint64 {
	return r.bo.Uint64(r.Read(offset, 8))
}
