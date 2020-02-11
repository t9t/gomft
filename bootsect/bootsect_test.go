package bootsect_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/bootsect"
)

func TestParse(t *testing.T) {
	b, err := hex.DecodeString("eb52904e5446532020202000020800000000000000f800003f00ff0000280300000000008000800010825b740000000000000c00000000000200000000000000f600000001000000a370d74c31115c3e00000000fa33c08ed0bc007cfb68c0071f1e686600cb88160e0066813e03004e5446537515b441bbaa55cd13720c81fb55aa7506f7c101007503e9dd001e83ec18681a00b4488a160e008bf4161fcd139f83c4189e581f72e13b060b0075dba30f00c12e0f00041e5a33dbb900202bc866ff06110003160f008ec2ff061600e84b002bc877efb800bbcd1a6623c0752d6681fb54435041752481f90201721e166807bb1668700e1668090066536653665516161668b80166610e07cd1a33c0bf2810b9d80ffcf3aae95f01909066601e0666a111006603061c001e66680000000066500653680100681000b4428a160e00161f8bf4cd1366595b5a665966591f0f82160066ff06110003160f008ec2ff0e160075bc071f6661c3a0f801e80900a0fb01e80300f4ebfdb4018bf0ac3c007409b40ebb0700cd10ebf2c30d0a41206469736b2072656164206572726f72206f63637572726564000d0a424f4f544d4752206973206d697373696e67000d0a424f4f544d475220697320636f6d70726573736564000d0a5072657373204374726c2b416c742b44656c20746f20726573746172740d0a008ca9bed6000055aa")
	require.Nilf(t, err, "unable to convert input hex to []byte: %v", err)

	ret, err := bootsect.Parse(b[0:80])
	require.Nilf(t, err, "could not parse boot sector: %v", err)
	expected := bootsect.BootSector{
		OemId:                        "NTFS    ",
		BytesPerSector:               512,
		SectorsPerCluster:            8,
		MediaDescriptor:              0xF8,
		SectorsPerTrack:              63,
		NumberofHeads:                255,
		HiddenSectors:                10240,
		TotalSectors:                 0x745b8210,
		MftClusterNumber:             0xc0000,
		MftMirrorClusterNumber:       0x2,
		FileRecordSegmentSizeInBytes: 1024,
		IndexBufferSizeInBytes:       4096,
		VolumeSerialNumber:           []byte{0xA3, 0x70, 0xD7, 0x4C, 0x31, 0x11, 0x5C, 0x3E},
	}

	assert.Equal(t, expected, ret)
}
