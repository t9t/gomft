package utf16

import (
	"encoding/binary"
	"unicode/utf16"
)

// Decode the input data as UTF-16 using the provided byte order and convert the result to a string. The input data
// length must be a multiple of 2. DecodeString will panic if that is not the case.
func DecodeString(b []byte, bo binary.ByteOrder) string {
	slen := len(b) / 2
	shorts := make([]uint16, slen)
	for i := 0; i < slen; i++ {
		bi := i * 2
		shorts[i] = bo.Uint16(b[bi : bi+2])
	}
	return string(utf16.Decode(shorts))
}
