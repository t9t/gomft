package bootsect_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/bootsect"
)

func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("test-bootsect.bin")
	require.Nilf(t, err, "unable to read test-bootsect.bin: %v", err)

	ret, err := bootsect.Parse(b[0:80])
	require.Nilf(t, err, "could not parse boot sector: %v", err)
	expected := bootsect.BootSector{
		OemId:                  "NTFS    ",
		BytesPerSector:         512,
		SectorsPerCluster:      8,
		MediaDescriptor:        0xF8,
		SectorsPerTrack:        63,
		NumberofHeads:          255,
		HiddenSectors:          10240,
		TotalSectors:           0x745b8210,
		MftClusterNumber:       0xc0000,
		MftMirrorClusterNumber: 0x2,
		FileRecordSegmentSize:  bootsect.BytesOrClusters{IsBytes: true, Value: 1024},
		IndexBufferSize:        bootsect.BytesOrClusters{IsBytes: false, Value: 1},
		VolumeSerialNumber:     []byte{0xA3, 0x70, 0xD7, 0x4C, 0x31, 0x11, 0x5C, 0x3E},
	}

	assert.Equal(t, expected, ret)
}
