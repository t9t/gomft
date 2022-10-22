package mft

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/t9t/gomft/binutil"
	"github.com/t9t/gomft/utf16"
)

var (
	reallyStrangeEpoch = time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
)

// StandardInformation represents the data contained in a $STANDARD_INFORMATION attribute.
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

// ParseStandardInformation parses the data of a $STANDARD_INFORMATION attribute's data (type
// AttributeTypeStandardInformation) into StandardInformation. Note that no additional correctness checks are done, so
// it's up to the caller to ensure the passed data actually represents a $STANDARD_INFORMATION attribute's data.
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

// FileAttribute represents a bit mask of various file attributes.
type FileAttribute uint32

// Bit values for FileAttribute. For example, a normal, hidden file has value 0x0082.
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

// Is checks if this FileAttribute's bit mask contains the specified attribute value.
func (a *FileAttribute) Is(c FileAttribute) bool {
	return *a&c == c
}

// FileNameNamespace indicates the namespace of a $FILE_NAME attribute's file name.
type FileNameNamespace byte

const (
	FileNameNamespacePosix    FileNameNamespace = 0
	FileNameNamespaceWin32    FileNameNamespace = 1
	FileNameNamespaceDos      FileNameNamespace = 2
	FileNameNamespaceWin32Dos FileNameNamespace = 3
)

// FileName represents the data of a $FILE_NAME attribute. ParentFileReference points to the MFT record that is the
// parent (ie. containing directory of this file). The AllocatedSize and ActualSize may be zero, in which case the file
// size may be found in a $DATA attribute instead (it could also be the ActualSize is zero, while the AllocatedSize does
// contain a value).
type FileName struct {
	ParentFileReference FileReference
	Creation            time.Time
	FileLastModified    time.Time
	MftLastModified     time.Time
	LastAccess          time.Time
	AllocatedSize       uint64
	ActualSize          uint64
	Flags               FileAttribute
	ExtendedData        uint32
	Namespace           FileNameNamespace
	Name                string
}

// ParseFileName parses the data of a $FILE_NAME attribute's data (type AttributeTypeFileName) into FileName. Note that
// no additional correctness checks are done, so it's up to the caller to ensure the passed data actually represents a
// $FILE_NAME attribute's data.
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
		ActualSize:          r.Uint64(0x30),
		Flags:               FileAttribute(r.Uint32(0x38)),
		ExtendedData:        r.Uint32(0x3c),
		Namespace:           FileNameNamespace(r.Byte(0x41)),
		Name:                utf16.DecodeString(r.Read(0x42, fileNameLength), binary.LittleEndian),
	}, nil
}

// AttributeListEntry represents an entry in an $ATTRIBUTE_LIST attribute. The Type indicates the attribute type, while
// the BaseRecordReference indicates which MFT record the attribute is located in (ie. an "extension record", if it is
// not the same as the one where the $ATTRIBUTE_LIST is located).
type AttributeListEntry struct {
	Type                AttributeType
	Name                string
	StartingVCN         uint64
	BaseRecordReference FileReference
	AttributeId         uint16
}

// ParseAttributeList parses the data of a $ATTRIBUTE_LIST attribute's data (type AttributeTypeAttributeList) into a
// list of AttributeListEntry. Note that no additional correctness checks are done, so it's up to the caller to ensure
// the passed data actually represents a $ATTRIBUTE_LIST attribute's data.
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
			name = utf16.DecodeString(r.Read(nameOffset, nameLength*2), binary.LittleEndian)
		}
		baseRef, err := ParseFileReference(r.Read(0x10, 8))
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

// CollationType indicates how the entries in an index should be ordered.
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

// IndexRoot represents the data (header and entries) of an $INDEX_ROOT attribute, which typically is the root of a
// directory's B+tree index containing file names of the directory (but could be use for other types of indices, too).
// The AttributeType is the type of attributes that are contained in the entries (currently only $FILE_NAME attributes
// are supported).
type IndexRoot struct {
	AttributeType     AttributeType
	CollationType     CollationType
	BytesPerRecord    uint32
	ClustersPerRecord uint32
	Flags             uint32
	Entries           []IndexEntry
}

// IndexEntry represents an entry in an B+tree index. Currently only $FILE_NAME attribute entries are supported. The
// FileReference points to the MFT record of the indexed file.
type IndexEntry struct {
	FileReference FileReference
	Flags         uint32
	FileName      FileName
	SubNodeVCN    uint64
}

// IndexBlock represents an IndexHeader preceding IndexEntry data. The EntryOffset defines the beginning of the
// first IndexEntry relative to the position of EntryOffset at 0x18.
// http://inform.pucp.edu.pe/~inf232/Ntfs/ntfs_doc_v0.5/concepts/index_header.html
type IndexBlock struct {
	Signature            string
	UpdateSequenceOffset uint16
	UpdateSequenceSize   uint16
	UpdateSequenceNumber uint16
	LSN                  uint64 // $LogFile Sequence Number
	EntryOffset          uint32
	TotalEntrySize       uint32
	AllocEntrySize       uint32
	NotLeaf              byte
}

// ParseIndexRoot parses the data of a $INDEX_ROOT attribute's data (type AttributeTypeIndexRoot) into
// IndexRoot. Note that no additional correctness checks are done, so it's up to the caller to ensure the passed data
// actually represents a $INDEX_ROOT attribute's data.
func ParseIndexRoot(b []byte) (IndexRoot, error) {
	if len(b) < 32 {
		return IndexRoot{}, fmt.Errorf("expected at least %d bytes but got %d", 32, len(b))
	}
	r := binutil.NewLittleEndianReader(b)
	attributeType := AttributeType(r.Uint32(0x00))
	if attributeType != AttributeTypeFileName {
		return IndexRoot{}, fmt.Errorf("unable to handle attribute type %d (%s) in $INDEX_ROOT", attributeType, attributeType.Name())
	}

	uTotalSize := r.Uint32(0x14)
	if int64(uTotalSize) > maxInt {
		return IndexRoot{}, fmt.Errorf("index root size %d overflows maximum int value %d", uTotalSize, maxInt)
	}
	totalSize := int(uTotalSize)
	expectedSize := totalSize + 16
	if len(b) < expectedSize {
		return IndexRoot{}, fmt.Errorf("expected %d bytes in $INDEX_ROOT but is %d", expectedSize, len(b))
	}
	entries := []IndexEntry{}
	if totalSize >= 16 {
		parsed, err := ParseIndexEntries(r.Read(0x20, totalSize-16))
		if err != nil {
			return IndexRoot{}, fmt.Errorf("error parsing index entries: %v", err)
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

// ParseIndexBlock parses the data of a $INDEX_ALLOCATION attribute into IndexBlock.
// Note that no additional correctness checks are done, so it's up to the caller to ensure the passed data
// actually represents a $INDEX_ALLOCATION attribute's data.
func ParseIndexBlock(b []byte) (IndexBlock, error) {
	if len(b) < 36 {
		return IndexBlock{}, fmt.Errorf("expected at least %d bytes but got %d", 36, len(b))
	}

	r := binutil.NewLittleEndianReader(b)
	signature := string(r.Read(0x00, 0x04))
	sequenceNumberOffset := r.Uint16(0x04)
	sequenceNumberSize := r.Uint16(0x06)
	updateSequenceNumber := r.Uint16(int(sequenceNumberOffset))
	lsn := r.Uint64(0x08)

	entryOffset := r.Uint32(0x18)
	totalEntrySize := r.Uint32(0x1C)
	allocEntrySize := r.Uint32(0x20)
	notLeaf := r.Read(0x24, 1)[0]

	return IndexBlock{Signature: signature,
		UpdateSequenceOffset: sequenceNumberOffset,
		UpdateSequenceSize:   sequenceNumberSize,
		UpdateSequenceNumber: updateSequenceNumber,
		LSN:                  lsn, // $LogFile Sequence Number
		EntryOffset:          entryOffset,
		TotalEntrySize:       totalEntrySize,
		AllocEntrySize:       allocEntrySize,
		NotLeaf:              notLeaf}, nil
}

// ParseIndexEntries parses the given raw bytes into a list of IndexEntry objects.
func ParseIndexEntries(b []byte) ([]IndexEntry, error) {
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
				return entries, fmt.Errorf("error parsing $FILE_NAME record in index entry: %v", err)
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
		if isLastEntryInNode {
			break
		}
	}
	return entries, nil
}

// ConvertFileTime converts a Windows "file time" to a time.Time. A "file time" is a 64-bit value that represents the
// number of 100-nanosecond intervals that have elapsed since 12:00 A.M. January 1, 1601 Coordinated Universal Time
// (UTC). See also: https://docs.microsoft.com/en-us/windows/win32/sysinfo/file-times
func ConvertFileTime(timeValue uint64) time.Time {
	dur := time.Duration(int64(timeValue))
	r := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100; i++ {
		r = r.Add(dur)
	}
	return r
}
