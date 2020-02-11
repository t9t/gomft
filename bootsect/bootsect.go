/*
	Package bootsect provides functions to parse the boot sector (also sometimes called Volume Boot Record, VBR, or
	$Boot file) of an NTFS volume.
*/
package bootsect

import (
	"fmt"

	"github.com/t9t/gomft/binutil"
)

// BootSector represents the parsed data of an NTFS boot sector. The OemId should typically be "NTFS    " ("NTFS"
// followed by 4 trailing spaces) for a valid NTFS boot sector.
type BootSector struct {
	OemId                        string
	BytesPerSector               int
	SectorsPerCluster            int
	MediaDescriptor              byte
	SectorsPerTrack              int
	NumberofHeads                int
	HiddenSectors                int
	TotalSectors                 uint64
	MftClusterNumber             uint64
	MftMirrorClusterNumber       uint64
	FileRecordSegmentSizeInBytes int
	IndexBufferSizeInBytes       int
	VolumeSerialNumber           []byte
}

// Parse parses the data of an NTFS boot sector into a BootSector structure.
func Parse(data []byte) (BootSector, error) {
	if len(data) < 80 {
		return BootSector{}, fmt.Errorf("boot sector data should be at least 80 bytes but is %d", len(data))
	}
	r := binutil.NewLittleEndianReader(data)
	bytesPerSector := int(r.Uint16(0x0B))
	sectorsPerCluster := int(int8(r.Byte(0x0D)))
	if sectorsPerCluster < 0 {
		// Quoth Wikipedia: The number of sectors in a cluster. If the value is negative, the amount of sectors is 2
		// to the power of the absolute value of this field.
		sectorsPerCluster = 1 << -sectorsPerCluster
	}
	bytesPerCluster := bytesPerSector * sectorsPerCluster
	return BootSector{
		OemId:                        string(r.Read(0x03, 8)),
		BytesPerSector:               bytesPerSector,
		SectorsPerCluster:            sectorsPerCluster,
		MediaDescriptor:              r.Byte(0x15),
		SectorsPerTrack:              int(r.Uint16(0x18)),
		NumberofHeads:                int(r.Uint16(0x1A)),
		HiddenSectors:                int(r.Uint16(0x1C)),
		TotalSectors:                 r.Uint64(0x28),
		MftClusterNumber:             r.Uint64(0x30),
		MftMirrorClusterNumber:       r.Uint64(0x38),
		FileRecordSegmentSizeInBytes: bytesOrClustersToBytes(r.Byte(0x40), bytesPerCluster),
		IndexBufferSizeInBytes:       bytesOrClustersToBytes(r.Byte(0x44), bytesPerCluster),
		VolumeSerialNumber:           binutil.Duplicate(r.Read(0x48, 8)),
	}, nil
}

func bytesOrClustersToBytes(b byte, bytesPerCluster int) int {
	// From Wikipedia:
	// A positive value denotes the number of clusters in a File Record Segment. A negative value denotes the amount of
	// bytes in a File Record Segment, in which case the size is 2 to the power of the absolute value.
	// (0xF6 = -10 → 210 = 1024).
	i := int(int8(b))
	if i < 0 {
		return 1 << -i
	}
	return i * bytesPerCluster
}
