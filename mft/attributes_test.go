package mft_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/mft"
)

func TestFileAttribute(t *testing.T) {
	a := mft.FileAttribute(0x83)

	// just a sample
	assert.True(t, a.Is(mft.FileAttributeReadOnly))
	assert.True(t, a.Is(mft.FileAttributeHidden))
	assert.True(t, a.Is(mft.FileAttributeNormal))
	assert.False(t, a.Is(mft.FileAttributeDevice))
	assert.False(t, a.Is(mft.FileAttributeCompressed))
}

func TestParseStandardInformation(t *testing.T) {
	input := decodeHex(t, "8d07703c89d7d5018d07703c89d6d5018d07703c89d6d5018d07703c89d6d501200000000000A30005000000010000000070000001100000000010000000000028820f4b05000000")
	out, err := mft.ParseStandardInformation(input)
	require.Nilf(t, err, "could not parse attribute: %v", err)
	expected := mft.StandardInformation{
		Creation:                time.Date(2020, time.January, 30, 16, 20, 50, 176398100, time.UTC),
		FileLastModified:        time.Date(2020, time.January, 29, 9, 48, 19, 13620500, time.UTC),
		MftLastModified:         time.Date(2020, time.January, 29, 9, 48, 19, 13620500, time.UTC),
		LastAccess:              time.Date(2020, time.January, 29, 9, 48, 19, 13620500, time.UTC),
		FileAttributes:          mft.FileAttribute(32),
		MaximumNumberOfVersions: 10682368,
		VersionNumber:           5,
		ClassId:                 1,
		OwnerId:                 28672,
		SecurityId:              4097,
		QuotaCharged:            1048576,
		UpdateSequenceNumber:    22734144040,
	}
	assert.Equal(t, expected, out)
}

func TestParseFileName(t *testing.T) {
	input := decodeHex(t, "e2680900000004007064eacc62b2d501000f014577c1cf01808beacc62b2d5017064eacc62b2d50100a00100000000002a9801000000000020000000000000000c036c006f0067006f002d003200350030002e0070006e006700")
	out, err := mft.ParseFileName(input)
	require.Nilf(t, err, "could not parse attribute: %v", err)
	expected := mft.FileName{
		ParentFileReference: mft.FileReference{RecordNumber: 616674, SequenceNumber: 4},
		Creation:            time.Date(2019, time.December, 14, 9, 42, 29, 175000000, time.UTC),
		FileLastModified:    time.Date(2014, time.August, 26, 21, 47, 02, 0, time.UTC),
		MftLastModified:     time.Date(2019, time.December, 14, 9, 42, 29, 176000000, time.UTC),
		LastAccess:          time.Date(2019, time.December, 14, 9, 42, 29, 175000000, time.UTC),
		AllocatedSize:       106496,
		ActualSize:          104490,
		Flags:               mft.FileAttribute(32),
		ExtendedData:        0,
		Namespace:           3,
		Name:                "logo-250.png",
	}
	assert.Equal(t, expected, out)
}

func TestParseAttributeList(t *testing.T) {
	input := decodeHex(t, "100000002000001a00000000000000003b410500000009000000444300000000300000002000001a00000000000000003b410500000009000500000000000000800000002000001a00000000000000004e1905000000a9000000000000000000800000002000001abaec01000000000052400500000049000000000000000000800000002000001ab7180300000000000241050000000f000000000000000000800000002000001a103e0400000000000941050000001d000000000000000000")
	out, err := mft.ParseAttributeList(input)
	require.Nilf(t, err, "could not parse attribute: %v", err)

	expected := []mft.AttributeListEntry{
		mft.AttributeListEntry{Type: mft.AttributeTypeStandardInformation, BaseRecordReference: mft.FileReference{RecordNumber: 344379, SequenceNumber: 9}},
		mft.AttributeListEntry{Type: mft.AttributeTypeFileName, BaseRecordReference: mft.FileReference{RecordNumber: 344379, SequenceNumber: 9}, AttributeId: 5},
		mft.AttributeListEntry{Type: mft.AttributeTypeData, BaseRecordReference: mft.FileReference{RecordNumber: 334158, SequenceNumber: 169}},
		mft.AttributeListEntry{Type: mft.AttributeTypeData, StartingVCN: 0x1ecba, BaseRecordReference: mft.FileReference{RecordNumber: 344146, SequenceNumber: 73}},
		mft.AttributeListEntry{Type: mft.AttributeTypeData, StartingVCN: 0x318b7, BaseRecordReference: mft.FileReference{RecordNumber: 344322, SequenceNumber: 15}},
		mft.AttributeListEntry{Type: mft.AttributeTypeData, StartingVCN: 0x43e10, BaseRecordReference: mft.FileReference{RecordNumber: 344329, SequenceNumber: 29}},
	}
	assert.Equal(t, expected, out)
}

func TestParseIndexRoot(t *testing.T) {
	input := decodeHex(t, "30000000010000000010000001000000100000008800000088000000000000005fac0600000006006800520000000000398c060000003b00de3ef1e234dcd501de3ef1e234dcd50118dbd2e334dcd501de3ef1e234dcd501000000000000000000000000000000002000000000000000080374006500730074002e0074007800740000002800000000000000000000001000000002000000")
	out, err := mft.ParseIndexRoot(input)
	require.Nilf(t, err, "could not parse attribute: %v", err)

	expected := mft.IndexRoot{
		AttributeType:     mft.AttributeTypeFileName,
		CollationType:     1,
		BytesPerRecord:    4096,
		ClustersPerRecord: 1,
		Flags:             0,
		Entries: []mft.IndexEntry{
			mft.IndexEntry{
				FileReference: mft.FileReference{RecordNumber: 437343, SequenceNumber: 6},
				Flags:         0,
				FileName: mft.FileName{
					ParentFileReference: mft.FileReference{RecordNumber: 429113, SequenceNumber: 59},
					Creation:            time.Date(2020, time.February, 5, 14, 59, 38, 116886200, time.UTC),
					FileLastModified:    time.Date(2020, time.February, 5, 14, 59, 38, 116886200, time.UTC),
					MftLastModified:     time.Date(2020, time.February, 5, 14, 59, 39, 595445600, time.UTC),
					LastAccess:          time.Date(2020, time.February, 5, 14, 59, 38, 116886200, time.UTC),
					AllocatedSize:       0,
					ActualSize:          0,
					Flags:               32,
					ExtendedData:        0,
					Namespace:           3,
					Name:                "test.txt",
				},
				SubNodeVCN: 0x0,
			},
			mft.IndexEntry{FileReference: mft.FileReference{}, Flags: 2, FileName: mft.FileName{}, SubNodeVCN: 0x0},
		},
	}
	assert.Equal(t, expected, out)
}
