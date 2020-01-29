package bootsect_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/t9t/gomft/bootsect"
)

func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("bootsect.bin")
	if err != nil {
		t.Fatal("unable to read bootsect.bin:", err)
	}

	ret := bootsect.Parse(b)
	expected := bootsect.BootSector{
		OemId:                              "NTFS    ",
		BytesPerSector:                     512,
		SectorsPerCluster:                  8,
		MediaDescriptor:                    0xF8,
		SectorsPerTrack:                    63,
		NumberofHeads:                      255,
		HiddenSectors:                      10240,
		TotalSectors:                       0x745b8210,
		MftClusterNumber:                   0xc0000,
		MftMirrorClusterNumber:             0x2,
		BytesOrClusterPerFileSecordSegment: 0xF6,
		BytersOrClusterPerIndexBuffer:      0x1,
		VolumeSerialNumber:                 []byte{0xA3, 0x70, 0xD7, 0x4C, 0x31, 0x11, 0x5C, 0x3E},
	}

	assert.Equal(t, expected, ret)
}
