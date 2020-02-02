package main

import (
	"fmt"
	"io"
	"os"

	"github.com/t9t/gomft/bootsect"
	"github.com/t9t/gomft/fragment"
	"github.com/t9t/gomft/mft"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Will dump the MFT of a drive to a file. The drive should be NTFS formatted.")
		fmt.Println("\nUsage:")
		fmt.Printf("\t%s <drive> <output file>", os.Args[0])
		fmt.Println("\n\nFor example:")
		fmt.Printf("\t%s C: D:\\c.mft\n", os.Args[0])
		os.Exit(1)
		return
	}

	drive := `\\.\` + os.Args[1]
	outfile := os.Args[2]

	in, err := os.Open(drive)
	if err != nil {
		fatalf(2, "Unable to open drive using path %s: %v\n", drive, err)
	}
	defer in.Close()

	bootSectorData := make([]byte, 512)
	_, err = io.ReadFull(in, bootSectorData)
	if err != nil {
		fatalf(3, "Unable to read boot sector: %v\n", err)
	}

	bootSector, err := bootsect.Parse(bootSectorData)
	if err != nil {
		fatalf(4, "Unable to parse boot sector data: %v\n", err)
	}

	bytesPerCluster := bootSector.BytesPerSector * bootSector.SectorsPerCluster
	mftPosInBytes := int64(bootSector.MftClusterNumber) * int64(bytesPerCluster)

	_, err = in.Seek(mftPosInBytes, 0)
	if err != nil {
		fatalf(5, "Unable to seek to MFT position: %v\n", err)
	}

	mftData := make([]byte, bootSector.FileRecordSegmentSize.ToBytes(bytesPerCluster))
	_, err = io.ReadFull(in, mftData)
	if err != nil {
		fatalf(6, "Unable to read $MFT record: %v\n", err)
	}

	record, err := mft.ParseRecord(mftData)
	if err != nil {
		fatalf(7, "Unable to parse $MFT record: %v\n", err)
	}

	dataAttribute, found := findDataAttribute(record.Attributes)
	if !found {
		fatalf(8, "No $DATA attribute found in $MFT record\n")
	}

	if dataAttribute.Resident {
		fatalf(9, "Don't know how to handle resident $DATA attribute in $MFT record\n")
	}

	dataRuns, err := mft.ParseDataRuns(dataAttribute.Data)
	if err != nil {
		fatalf(10, "Unable to parse dataruns in $MFT $DATA record: %v\n", err)
	}
	if len(dataRuns) == 0 {
		fatalf(11, "No dataruns found in $MFT $DATA record\n")
	}

	fragments := mft.DataRunsToFragments(dataRuns, bytesPerCluster)
	totalLength := int64(0)
	for _, frag := range fragments {
		totalLength += int64(frag.Length)
	}

	out, err := createFileIfNotExist(outfile)
	if err != nil {
		fatalf(12, "Unable to open output file: %v\n", err)
	}
	defer out.Close()

	n, err := io.Copy(out, fragment.NewReader(in, fragments))
	if err != nil {
		fatalf(13, "Error copying data to output file: %v\n", err)
	}

	if n != totalLength {
		fatalf(14, "Expected to copy %d bytes, but copied only %d\n", totalLength, n)
	}
}

func createFileIfNotExist(outfile string) (*os.File, error) {
	return os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
}

func fatalf(exitCode int, format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(exitCode)
}

func findDataAttribute(attributes []mft.Attribute) (mft.Attribute, bool) {
	for _, attr := range attributes {
		if attr.Type == mft.ATTRIBUTE_TYPE_DATA {
			return attr, true
		}
	}

	return mft.Attribute{}, false
}
