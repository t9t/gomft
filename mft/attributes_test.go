package mft_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/mft"
)

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
		ParentFileReference: 1125899907459298,
		Creation:                time.Date(2019, time.December, 14, 9, 42, 29, 175000000, time.UTC),
		FileLastModified:        time.Date(2014, time.August, 26, 21, 47, 02, 0, time.UTC),
		MftLastModified:         time.Date(2019, time.December, 14, 9, 42, 29, 176000000, time.UTC),
		LastAccess:              time.Date(2019, time.December, 14, 9, 42, 29, 175000000, time.UTC),
		AllocatedSize:       106496,
		RealSize:            104490,
		Flags:               mft.FileAttribute(32),
		ExtendedData:        0,
		Namespace:           3,
		Name:                "logo-250.png",

	}
	assert.Equal(t, expected, out)
}
