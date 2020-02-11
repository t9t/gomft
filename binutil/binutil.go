// Package binutil contains some helpful utilities for reading binary data from byte slices.
package binutil

import "encoding/binary"

// Duplicate creates a full copy of the input byte slice.
func Duplicate(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

// IsOnlyZeroes return true when the input data is all bytes of zero value and false if any of the bytes has a nonzero
// value.
func IsOnlyZeroes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// BinReader helps to read data from a byte slice using an offset and a data length (instead two offsets when using
// a slice expression). For example b[2:4] yields the same as Read(2, 2) using a BinReader over b. Also some convenient
// methods are provided to read integer values using a binary.ByteOrder from the slice directly.
// 
// Note that methods that return a []byte may not necessarily copy the data, so modifying the returned slice may also
// affect the data in the BinReader.
//
// Methods will panic when any offset or length is outside of the bounds of the original data.
type BinReader struct {
	data []byte
	bo   binary.ByteOrder
}

// NewBinReader creates a BinReader over data using the specified binary.ByteOrder. The data slice is stored directly,
// no copy is made, so modifying the original slice will also affect the returned BinReader.
func NewBinReader(data []byte, bo binary.ByteOrder) *BinReader {
	return &BinReader{data: data, bo: bo}
}

// NewLittleEndianReader creates a BinReader over data using binary.LittleEndian. The data slice is stored directly,
// no copy is made, so modifying the original slice will also affect the returned BinReader.
func NewLittleEndianReader(data []byte) *BinReader {
	return NewBinReader(data, binary.LittleEndian)
}

// NewLittleEndianReader creates a BinReader over data using binary.BigEndian. The data slice is stored directly,
// no copy is made, so modifying the original slice will also affect the returned BinReader.
func NewBigEndianReader(data []byte, bo binary.ByteOrder) *BinReader {
	return NewBinReader(data, binary.BigEndian)
}

// Data returns all data inside this BinReader.
func (r *BinReader) Data() []byte {
	return r.data
}

// ByteOrder returns the ByteOrder for this BinReader.
func (r *BinReader) ByteOrder() binary.ByteOrder {
	return r.bo
}

// Length returns the length of the contained data.
func (r *BinReader) Length() int {
	return len(r.data)
}

// Read reads an amount of bytes as specified by length from the provided offset. The returned slice's length is the
// same as the specified length.
func (r *BinReader) Read(offset int, length int) []byte {
	return r.data[offset : offset+length]
}

// Reader returns a new BinReader over the data read by Read(offset, length) using the same ByteOrder as this reader.
// There is no guarantee a copy of the data is made, so modifying the new reader's data may affect the original.
func (r *BinReader) Reader(offset int, length int) *BinReader {
	return &BinReader{data: r.data[offset : offset+length], bo: r.bo}
}

// Byte returns the byte at the position indicated by the offset.
func (r *BinReader) Byte(offset int) byte {
	return r.Read(offset, 1)[0]
}

// ReadFrom returns all data starting at the specified offset.
func (r *BinReader) ReadFrom(offset int) []byte {
	return r.data[offset:]
}

// ReaderFrom returns a BinReader over the data read by ReadFrom(offset) using the same ByteOrder as this reader.
// There is no guarantee a copy of the data is made, so modifying the new reader's data may affect the original.
func (r *BinReader) ReaderFrom(offset int) *BinReader {
	return &BinReader{data: r.data[offset:], bo: r.bo}
}

// Uint16 reads 2 bytes from the provided offset and parses them into a uint16 using the provided ByteOrder.
func (r *BinReader) Uint16(offset int) uint16 {
	return r.bo.Uint16(r.Read(offset, 2))
}

// Uint32 reads 4 bytes from the provided offset and parses them into a uint32 using the provided ByteOrder.
func (r *BinReader) Uint32(offset int) uint32 {
	return r.bo.Uint32(r.Read(offset, 4))
}

// Uint64 reads 8 bytes from the provided offset and parses them into a uint64 using the provided ByteOrder.
func (r *BinReader) Uint64(offset int) uint64 {
	return r.bo.Uint64(r.Read(offset, 8))
}
