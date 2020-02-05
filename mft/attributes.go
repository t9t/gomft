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
	if len(b) < 72 {
		return StandardInformation{}, fmt.Errorf("expected at least %d bytes but got %d", 72, len(b))
	}

	r := binutil.NewLittleEndianReader(b)
	return StandardInformation{
		Creation:                convertFileTime(r.Uint64(0x00)),
		FileLastModified:        convertFileTime(r.Uint64(0x08)),
		MftLastModified:         convertFileTime(r.Uint64(0x10)),
		LastAccess:              convertFileTime(r.Uint64(0x18)),
		FileAttributes:          FileAttribute(r.Uint32(0x20)),
		MaximumNumberOfVersions: r.Uint32(0x24),
		VersionNumber:           r.Uint32(0x28),
		ClassId:                 r.Uint32(0x2C),
		OwnerId:                 r.Uint32(0x30),
		SecurityId:              r.Uint32(0x34),
		QuotaCharged:            r.Uint64(0x38),
		UpdateSequenceNumber:    r.Uint64(0x40),
	}, nil
}

type FileNameNamespace byte
type FileName struct {
	ParentFileReference uint64
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
	return FileName{
		ParentFileReference: r.Uint64(0x00),
		Creation:            convertFileTime(r.Uint64(0x08)),
		FileLastModified:    convertFileTime(r.Uint64(0x10)),
		MftLastModified:     convertFileTime(r.Uint64(0x18)),
		LastAccess:          convertFileTime(r.Uint64(0x20)),
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
		entry := AttributeListEntry{
			Type:                AttributeType(r.Uint32(0)),
			Name:                name,
			StartingVCN:         r.Uint64(0x08),
			BaseRecordReference: FileReference(binutil.Duplicate(r.Read(0x08, 8))),
			AttributeId:         r.Uint16(0x18),
		}
		entries = append(entries, entry)
		b = r.ReadFrom(entryLength)
	}
	return entries, nil
}

func convertFileTime(timeValue uint64) time.Time {
	dur := time.Duration(int64(timeValue))
	r := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		r = r.Add(dur)
	}
	return r
}
