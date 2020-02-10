package mft_test

import (
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/fragment"
	"github.com/t9t/gomft/mft"
)

func TestParseRecordHeader(t *testing.T) {
	b := readTestMft(t)
	header, err := mft.ParseRecordHeader(b[:42])
	require.Nilf(t, err, "could not parse record header: %v", err)
	expected := mft.RecordHeader{
		Signature:             []byte{'F', 'I', 'L', 'E'},
		UpdateSequenceOffset:  48,
		UpdateSequenceSize:    3,
		LogFileSequenceNumber: 25695988020,
		SequenceNumber:        145,
		HardLinkCount:         1,
		FirstAttributeOffset:  56,
		Flags:                 mft.RecordFlag(mft.RecordFlagInUse),
		ActualSize:            480,
		AllocatedSize:         1024,
		BaseRecordReference:   mft.FileReference{RecordNumber: 18446727447098470560, SequenceNumber: 36880},
		NextAttributeId:       8,
		RecordNumber:          0,
	}

	assert.Equal(t, expected, header)
}

func TestParseAttributes(t *testing.T) {
	b := readTestMft(t)
	attributeData := b[56:]
	attributes, err := mft.ParseAttributes(attributeData)
	require.Nilf(t, err, "error parsing attributes: %v", err)

	expectedAttributes := []mft.Attribute{
		mft.Attribute{Type: 16, Resident: true, Flags: 0, AttributeId: 0, Data: []byte{0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		mft.Attribute{Type: 48, Resident: true, Flags: 0, AttributeId: 3, Data: []byte{0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x94, 0xF0, 0x48, 0x96, 0x5B, 0x2F, 0xCC, 0x1, 0x0, 0x0, 0xBC, 0x39, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xBC, 0x39, 0x0, 0x0, 0x0, 0x0, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4, 0x3, 0x24, 0x0, 0x4D, 0x0, 0x46, 0x0, 0x54, 0x0}},
		mft.Attribute{Type: 128, Resident: false, Flags: 0, AttributeId: 1, AllocatedSize: 1920466944, ActualSize: 1920466944, Data: []byte{0x33, 0x20, 0xC8, 0x0, 0x0, 0x0, 0xC, 0x43, 0x22, 0xB5, 0x0, 0xBA, 0x5, 0x5C, 0x3, 0x43, 0x81, 0xDE, 0x0, 0x65, 0xCF, 0x47, 0x4, 0x43, 0x84, 0xB3, 0x0, 0x5D, 0x8B, 0xEF, 0x9, 0x43, 0xB0, 0xE1, 0x0, 0x90, 0xB4, 0xB5, 0x18, 0x43, 0x0, 0xC8, 0x0, 0xF4, 0xEA, 0x13, 0x1, 0x43, 0x6, 0xC8, 0x0, 0x9A, 0x3A, 0x5A, 0xFE, 0x43, 0x12, 0xC8, 0x0, 0xF4, 0x7, 0x4D, 0xFE, 0x33, 0xF, 0xC8, 0x0, 0x23, 0xD4, 0xC0, 0x42, 0x62, 0x16, 0x54, 0x2, 0x95, 0x3, 0x0, 0x0, 0x0}},
		mft.Attribute{Type: 176, Resident: false, Flags: 0, AttributeId: 7, AllocatedSize: 237568, ActualSize: 237024, Data: []byte{0x41, 0x3A, 0xBE, 0x84, 0x83, 0x0, 0x0, 0x0}},
	}

	assert.Equal(t, expectedAttributes, attributes)
}
func TestParseDataRuns(t *testing.T) {
	input := decodeHex(t, "3320c80000000c42e061a4b54507330dc8006fedb142365db3d89cfb32802b3a045b433d830054029301000000000000")

	runs, err := mft.ParseDataRuns(input)
	require.Nilf(t, err, "error parsing dataruns: %v", err)

	expected := []mft.DataRun{
		mft.DataRun{OffsetCluster: 786432, LengthInClusters: 51232},
		mft.DataRun{OffsetCluster: 122008996, LengthInClusters: 25056},
		mft.DataRun{OffsetCluster: -5116561, LengthInClusters: 51213},
		mft.DataRun{OffsetCluster: -73606989, LengthInClusters: 23862},
		mft.DataRun{OffsetCluster: 5964858, LengthInClusters: 11136},
		mft.DataRun{OffsetCluster: 26411604, LengthInClusters: 33597},
	}

	assert.Equal(t, expected, runs)
}

func TestDataRunsToFragments(t *testing.T) {
	runs := []mft.DataRun{
		mft.DataRun{OffsetCluster: 5521, LengthInClusters: 1337},
		mft.DataRun{OffsetCluster: -4408, LengthInClusters: 42},
		mft.DataRun{OffsetCluster: 7708, LengthInClusters: 13},
	}

	fragments := mft.DataRunsToFragments(runs, 512)
	expected := []fragment.Fragment{
		fragment.Fragment{Offset: 2826752, Length: 684544},
		fragment.Fragment{Offset: 569856, Length: 21504},
		fragment.Fragment{Offset: 4516352, Length: 6656},
	}

	assert.Equal(t, expected, fragments)
}

func TestParseAttributeNamedResidentAttribute(t *testing.T) {
	input := decodeHex(t, "8000000070000000000518000000050044000000280000002400530052004100540000000000000033ceb8f33800010310000c00040000000100000001000000000000000200000000000000000000000300000001000000000000000000000000000000f4c400000000000000000000")

	attribute, err := mft.ParseAttribute(input)
	require.Nilf(t, err, "error parsing attribute: %v", err)

	expected := mft.Attribute{Type: 0x80, Resident: true, Name: "$SRAT", Flags: 0, AttributeId: 5, Data: []byte{0x33, 0xce, 0xb8, 0xf3, 0x38, 0x0, 0x1, 0x3, 0x10, 0x0, 0xc, 0x0, 0x4, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf4, 0xc4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}}
	assert.Equal(t, expected, attribute)
}

func TestParseAttributeNamedNonResidentAttribute(t *testing.T) {
	input := decodeHex(t, "a000000050000000010440000000080000000000000000000200000000000000480000000000000000300000000000000030000000000000003000000000000024004900330030002103081200000000")

	attribute, err := mft.ParseAttribute(input)
	require.Nilf(t, err, "error parsing attribute: %v", err)

	expected := mft.Attribute{Type: 0xA0, Resident: false, Name: "$I30", Flags: 0, AttributeId: 8, AllocatedSize: 12288, ActualSize: 12288, Data: []byte{0x21, 0x3, 0x8, 0x12, 0x0, 0x0, 0x0, 0x0}}
	assert.Equal(t, expected, attribute)
}

func TestParseRecordFixup(t *testing.T) {
	input := decodeHex(t, "46494c4530000300755762ef19000000150002003800010098020000000400000000000000000000060000002a0000000c000000000000001000000060000000000000000000000048000000180000007e31192b21d6d50186468bb40eded4012e7d4e954dcbd5016c7f192b21d6d5012000040000000000000000000000000000000000161300000000000000000000a068d14a05000000300000007800000000000000000003005a000000180001003b000000000009007e31192b21d6d5017e31192b21d6d5017e31192b21d6d5017e31192b21d6d5010020040000000000000000000000000020000000000000000c0249004e0054004c00500052007e0031002e0044004c004c000000000000003000000080000000000000000000020062000000180001003b000000000009007e31192b21d6d5017e31192b21d6d5017e31192b21d6d5017e31192b21d6d501002004000000000000000000000000002000000000000000100149006e0074006c00500072006f00760069006400650072002e0064006c006c00000000000000800000004800000001000000000001000000000000000000410000000000000040000000000000000020040000000000381704000000000038170400000000004142f46ea0000000d00000002000000000000000000004000800000018000000780000007c000000e000000098000c0000000000000005007c000000180000007c000000000f64002443492e434154414c4f4748494e5400010060004d6963726f736f66742d57696e646f77732d436c69656e742d4465736b746f702d52657175697265642d5061636b616765303431367e333162663338353661643336346533357e616d6436347e7e31302e302e31383336322e3539322e63617400000000ffffffff82794711000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c00")

	_, err := mft.ParseRecord(input)
	require.Nilf(t, err, "error parsing attribute: %v", err)

	// without fixup, this record returns an error parsing attributes; no further assertions necessary
}

func TestParseFileReference(t *testing.T) {
	ref, err := mft.ParseFileReference([]byte{26, 179, 6, 0, 0, 0, 45, 0})
	require.Nilf(t, err, "error parsing reference: %v", err)
	expected := mft.FileReference{RecordNumber: 439066, SequenceNumber: 45}
	assert.Equal(t, expected, ref)
}

func readTestMft(t *testing.T) []byte {
	b, err := ioutil.ReadFile("test-mft.bin")
	require.Nilf(t, err, "unable to read test-mft.bin: %v", err)
	return b
}

func decodeHex(t *testing.T, s string) []byte {
	input, err := hex.DecodeString(s)
	require.Nilf(t, err, "unable to convert input hex to []byte: %v", err)
	return input
}

func TestRecordFlag(t *testing.T) {
	f := mft.RecordFlag(0)
	assert.False(t, f.Is(mft.RecordFlagInUse))
	assert.False(t, f.Is(mft.RecordFlagIsDirectory))
	assert.False(t, f.Is(mft.RecordFlagInExtend))
	assert.False(t, f.Is(mft.RecordFlagIsIndex))

	f = mft.RecordFlag(1)
	assert.True(t, f.Is(mft.RecordFlagInUse))
	assert.False(t, f.Is(mft.RecordFlagIsDirectory))
	assert.False(t, f.Is(mft.RecordFlagInExtend))
	assert.False(t, f.Is(mft.RecordFlagIsIndex))

	f = mft.RecordFlag(3)
	assert.True(t, f.Is(mft.RecordFlagInUse))
	assert.True(t, f.Is(mft.RecordFlagIsDirectory))
	assert.False(t, f.Is(mft.RecordFlagInExtend))
	assert.False(t, f.Is(mft.RecordFlagIsIndex))

	f = mft.RecordFlag(15)
	assert.True(t, f.Is(mft.RecordFlagInUse))
	assert.True(t, f.Is(mft.RecordFlagIsDirectory))
	assert.True(t, f.Is(mft.RecordFlagInExtend))
	assert.True(t, f.Is(mft.RecordFlagIsIndex))
}
