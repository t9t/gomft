package mft

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/t9t/gomft/binutil"
	"github.com/t9t/gomft/utf16"
)

type FileAttribute uint32

const (
	FileAttributeReadOnly          FileAttribute = 0x0001
	FileAttributeHidden            FileAttribute = 0x0002
	FileAttributeSystem            FileAttribute = 0x0004
	FileAttributeArchive           FileAttribute = 0x0020
	FileAttributeDevice            FileAttribute = 0x0040
	FileAttributeNormal            FileAttribute = 0x0080
	FileAttributeTemporary         FileAttribute = 0x0100
	FileAttributeSparseFile        FileAttribute = 0x0200
	FileAttributeReparsePoint      FileAttribute = 0x0400
	FileAttributeCompressed        FileAttribute = 0x1000
	FileAttributeOffline           FileAttribute = 0x1000
	FileAttributeNotContentIndexed FileAttribute = 0x2000
	FileAttributeEncrypted         FileAttribute = 0x4000
)

var (
	reallyStrangeEpoch = time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type StandardInformation struct {
	Creation                time.Time
	FileLastModified        time.Time
	MftLastModified         time.Time
	LastAccess              time.Time
	FileAttributes          FileAttribute
	MaximumNumberOfVersions uint32
	VersionNumber           uint32
	ClassId                 uint32
	OwnerId                 uint32
	SecurityId              uint32
	QuotaCharged            uint64
	UpdateSequenceNumber    uint64
}

func ParseStandardInformation(b []byte) (StandardInformation, error) {
	if len(b) < 48 {
		return StandardInformation{}, fmt.Errorf("expected at least %d bytes but got %d", 48, len(b))
	}

	r := binutil.NewLittleEndianReader(b)
	ownerId := uint32(0)
	securityId := uint32(0)
	quotaCharged := uint64(0)
	updateSequenceNumber := uint64(0)
	if len(b) >= 0x30+4 {
		ownerId = r.Uint32(0x30)
	}
	if len(b) >= 0x34+4 {
		securityId = r.Uint32(0x34)
	}
	if len(b) >= 0x38+8 {
		quotaCharged = r.Uint64(0x38)
	}
	if len(b) >= 0x40+8 {
		updateSequenceNumber = r.Uint64(0x40)
	}
	return StandardInformation{
		Creation:                ConvertFileTime(r.Uint64(0x00)),
		FileLastModified:        ConvertFileTime(r.Uint64(0x08)),
		MftLastModified:         ConvertFileTime(r.Uint64(0x10)),
		LastAccess:              ConvertFileTime(r.Uint64(0x18)),
		FileAttributes:          FileAttribute(r.Uint32(0x20)),
		MaximumNumberOfVersions: r.Uint32(0x24),
		VersionNumber:           r.Uint32(0x28),
		ClassId:                 r.Uint32(0x2C),
		OwnerId:                 ownerId,
		SecurityId:              securityId,
		QuotaCharged:            quotaCharged,
		UpdateSequenceNumber:    updateSequenceNumber,
	}, nil
}

type FileNameNamespace byte
type FileName struct {
	ParentFileReference FileReference
	Creation            time.Time
	FileLastModified    time.Time
	MftLastModified     time.Time
	LastAccess          time.Time
	AllocatedSize       uint64
	RealSize            uint64
	Flags               FileAttribute
	ExtendedData        uint32
	Namespace           FileNameNamespace
	Name                string
}

func ParseFileName(b []byte) (FileName, error) {
	if len(b) < 66 {
		return FileName{}, fmt.Errorf("expected at least %d bytes but got %d", 66, len(b))
	}

	fileNameLength := int(b[0x40 : 0x40+1][0]) * 2
	minExpectedSize := 66 + fileNameLength
	if len(b) < minExpectedSize {
		return FileName{}, fmt.Errorf("expected at least %d bytes but got %d", minExpectedSize, len(b))
	}

	r := binutil.NewLittleEndianReader(b)
	name, err := utf16.DecodeString(r.Read(0x42, fileNameLength), binary.LittleEndian)
	if err != nil {
		return FileName{}, fmt.Errorf("unable to decode file name: %w", err)
	}
	parentRef, err := ParseFileReference(r.Read(0x00, 8))
	if err != nil {
		return FileName{}, fmt.Errorf("unable to parse file reference: %v", err)
	}
	return FileName{
		ParentFileReference: parentRef,
		Creation:            ConvertFileTime(r.Uint64(0x08)),
		FileLastModified:    ConvertFileTime(r.Uint64(0x10)),
		MftLastModified:     ConvertFileTime(r.Uint64(0x18)),
		LastAccess:          ConvertFileTime(r.Uint64(0x20)),
		AllocatedSize:       r.Uint64(0x28),
		RealSize:            r.Uint64(0x30),
		Flags:               FileAttribute(r.Uint32(0x38)),
		ExtendedData:        r.Uint32(0x3c),
		Namespace:           FileNameNamespace(r.Byte(0x41)),
		Name:                name,
	}, nil
}

type AttributeListEntry struct {
	Type                AttributeType
	Name                string
	StartingVCN         uint64
	BaseRecordReference FileReference
	AttributeId         uint16
}

func ParseAttributeList(b []byte) ([]AttributeListEntry, error) {
	if len(b) < 26 {
		return []AttributeListEntry{}, fmt.Errorf("expected at least %d bytes but got %d", 26, len(b))
	}

	entries := make([]AttributeListEntry, 0)

	for len(b) > 0 {
		r := binutil.NewLittleEndianReader(b)
		entryLength := int(r.Uint16(0x04))
		if len(b) < entryLength {
			return entries, fmt.Errorf("expected at least %d bytes remaining for AttributeList entry but is %d", entryLength, len(b))
		}
		nameLength := int(r.Byte(0x06))
		name := ""
		if nameLength != 0 {
			nameOffset := int(r.Byte(0x07))
			parsed, err := utf16.DecodeString(r.Read(nameOffset, nameLength*2), binary.LittleEndian)
			if err != nil {
				return entries, fmt.Errorf("unable to parsed attribute name: %w", err)
			}
			name = parsed
		}
		baseRef, err := ParseFileReference(r.Read(0x08, 8))
		if err != nil {
			return entries, fmt.Errorf("unable to parse base record reference: %v", err)
		}
		entry := AttributeListEntry{
			Type:                AttributeType(r.Uint32(0)),
			Name:                name,
			StartingVCN:         r.Uint64(0x08),
			BaseRecordReference: baseRef,
			AttributeId:         r.Uint16(0x18),
		}
		entries = append(entries, entry)
		b = r.ReadFrom(entryLength)
	}
	return entries, nil
}

type CollationType uint32

const (
	CollationTypeBinary            CollationType = 0x00000000
	CollationTypeFileName          CollationType = 0x00000001
	CollationTypeUnicodeString     CollationType = 0x00000002
	CollationTypeNtofsULong        CollationType = 0x00000010
	CollationTypeNtofsSid          CollationType = 0x00000011
	CollationTypeNtofsSecurityHash CollationType = 0x00000012
	CollationTypeNtofsUlongs       CollationType = 0x00000013
)

type IndexRoot struct {
	AttributeType     AttributeType
	CollationType     CollationType
	BytesPerRecord    uint32
	ClustersPerRecord uint32
	Flags             uint32
	Entries           []IndexEntry
}

func ParseIndexRoot(b []byte) (IndexRoot, error) {
	if len(b) < 32 {
		return IndexRoot{}, fmt.Errorf("expected at least %d bytes but got %d", 32, len(b))
	}
	r := binutil.NewLittleEndianReader(b)
	attributeType := AttributeType(r.Uint32(0x00))
	if attributeType != AttributeTypeFileName {
		return IndexRoot{}, fmt.Errorf("unable to handle attribute type %d (%s) in $INDEX_ROOT", attributeType, attributeType.Name())
	}

	totalSize := int(r.Uint32(0x14))
	expectedSize := totalSize + 16
	if len(b) < expectedSize {
		return IndexRoot{}, fmt.Errorf("expected %d bytes in $INDEX_ROOT but is %d", expectedSize, len(b))
	}
	entries := []IndexEntry{}
	if totalSize >= 16 {
		parsed, err := parseIndexEntries(r.Read(0x20, totalSize-16))
		if err != nil {
			return IndexRoot{}, fmt.Errorf("error parsing index entries: %w", err)
		}
		entries = parsed
	}

	return IndexRoot{
		AttributeType:     attributeType,
		CollationType:     CollationType(r.Uint32(0x04)),
		BytesPerRecord:    r.Uint32(0x08),
		ClustersPerRecord: r.Uint32(0x0C),
		Flags:             r.Uint32(0x1C),
		Entries:           entries,
	}, nil
}

type IndexEntry struct {
	FileReference FileReference
	Flags         uint32
	FileName      FileName
	SubNodeVCN    uint64
}

func parseIndexEntries(b []byte) ([]IndexEntry, error) {
	if len(b) < 13 {
		return []IndexEntry{}, fmt.Errorf("expected at least %d bytes but got %d", 13, len(b))
	}
	entries := make([]IndexEntry, 0)
	for len(b) > 0 {
		r := binutil.NewLittleEndianReader(b)
		entryLength := int(r.Uint16(0x08))

		if len(b) < entryLength {
			return entries, fmt.Errorf("index entry length indicates %d bytes but got %d", entryLength, len(b))
		}

		flags := r.Uint32(0x0C)
		pointsToSubNode := flags&0b1 != 0
		isLastEntryInNode := flags&0b10 != 0
		contentLength := int(r.Uint16(0x0A))

		fileName := FileName{}
		if contentLength != 0 && !isLastEntryInNode {
			parsedFileName, err := ParseFileName(r.Read(0x10, contentLength))
			if err != nil {
				return entries, fmt.Errorf("error parsing $FILE_NAME record in index entry: %w", err)
			}
			fileName = parsedFileName
		}
		subNodeVcn := uint64(0)
		if pointsToSubNode {
			subNodeVcn = r.Uint64(entryLength - 8)
		}

		fileReference, err := ParseFileReference(r.Read(0x00, 8))
		if err != nil {
			return entries, fmt.Errorf("unable to file reference: %v", err)
		}
		entry := IndexEntry{
			FileReference: fileReference,
			Flags:         flags,
			FileName:      fileName,
			SubNodeVCN:    subNodeVcn,
		}
		entries = append(entries, entry)
		b = r.ReadFrom(entryLength)
	}
	return entries, nil
}

func ConvertFileTime(timeValue uint64) time.Time {
	dur := time.Duration(int64(timeValue))
	r := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		r = r.Add(dur)
	}
	return r
}
