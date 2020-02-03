package utf16

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

func DecodeString(b []byte, bo binary.ByteOrder) (string, error) {
	blen := len(b)
	if blen%2 != 0 {
		return "", errors.New("input data must have even number of bytes")
	}
	slen := blen / 2
	shorts := make([]uint16, slen)
	for i := 0; i < slen; i++ {
		bi := i * 2
		shorts[i] = bo.Uint16(b[bi : bi+2])
	}
	return string(utf16.Decode(shorts)), nil
}
