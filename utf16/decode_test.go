package utf16_test

import (
	"testing"

	"encoding/binary"
	"encoding/hex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/utf16"
)

func TestDecodeString_LittleEndian(t *testing.T) {
	input, err := hex.DecodeString("480065006c006c006f002c00200077006f0072006c00640020003dd84cdc")
	require.Nilf(t, err, "unable to convert input hex to []byte: %v", err)
	output := utf16.DecodeString(input, binary.LittleEndian)
	assert.Equal(t, "Hello, world 👌", output)
}

func TestDecodeString_BigEndian(t *testing.T) {
	input, err := hex.DecodeString("00480065006c006c006f002c00200077006f0072006c00640020d83ddc4c")
	require.Nilf(t, err, "unable to convert input hex to []byte: %v", err)
	output := utf16.DecodeString(input, binary.BigEndian)
	assert.Equal(t, "Hello, world 👌", output)
}
