package mft_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t9t/gomft/mft"
)

func TestParseRecordHeader(t *testing.T) {
	b, err := ioutil.ReadFile("test-mft.bin")
	if err != nil {
		t.Fatal("unable to read test-bootsect.bin:", err)
	}

	ret := mft.ParseRecordHeader(b)
	expected := mft.RecordHeader{
		Signature:             []byte{'F', 'I', 'L', 'E'},
		UpdateSequenceOffset:  48,
		UpdateSequenceSize:    3,
		LogFileSequenceNumber: 25695988020,
		RecordUsageNumber:     145,
		HardLinkCount:         1,
		FirstAttributeOffset:  56,
		Flags:                 []byte{0x01, 0x00},
		ActualSize:            480,
		AllocatedSize:         1024,
		BaseRecordReference:   []byte{0xA0, 0xB0, 0xC0, 0xD0, 0xE0, 0xF0, 0x10, 0x90},
		NextAttributeId:       8,
	}

	assert.Equal(t, expected, ret)
}
