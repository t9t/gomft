package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/t9t/gomft/bootsect"
	"github.com/t9t/gomft/fragment"
	"github.com/t9t/gomft/mft"
)

const supportedOemId = "NTFS    "

const (
	exitCodeUserError int = iota + 1
	exitCodeFunctionalError
	exitCodeTechnicalError
)

func main() {
	isWin := runtime.GOOS == "windows"
	if len(os.Args) != 3 {
		fmt.Println("Will dump the MFT of a volume to a file. The volume should be NTFS formatted.")
		fmt.Println("\nUsage:")
		fmt.Printf("\t%s <volume> <output file>", os.Args[0])
		fmt.Println("\n\nFor example:")
		if isWin {
			fmt.Printf("\t%s C: D:\\c.mft\n", os.Args[0])
		} else {
			fmt.Printf("\t%s /dev/sdb1 ~/sdb1.mft\n", os.Args[0])
		}
		os.Exit(exitCodeUserError)
		return
	}

	volume := os.Args[1]
	if isWin {
		volume = `\\.\` + volume
	}
	outfile := os.Args[2]

	in, err := os.Open(volume)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to open volume using path %s: %v\n", volume, err)
	}
	defer in.Close()

	bootSectorData := make([]byte, 512)
	_, err = io.ReadFull(in, bootSectorData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to read boot sector: %v\n", err)
	}

	bootSector, err := bootsect.Parse(bootSectorData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to parse boot sector data: %v\n", err)
	}

	if bootSector.OemId != supportedOemId {
		fatalf(exitCodeFunctionalError, "Unknown OemId (file system type) %q (expected %q)\n", bootSector.OemId, supportedOemId)
	}

	bytesPerCluster := bootSector.BytesPerSector * bootSector.SectorsPerCluster
	mftPosInBytes := int64(bootSector.MftClusterNumber) * int64(bytesPerCluster)

	_, err = in.Seek(mftPosInBytes, 0)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to seek to MFT position: %v\n", err)
	}

	mftData := make([]byte, bootSector.FileRecordSegmentSize.ToBytes(bytesPerCluster))
	_, err = io.ReadFull(in, mftData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to read $MFT record: %v\n", err)
	}

	record, err := mft.ParseRecord(mftData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to parse $MFT record: %v\n", err)
	}

	dataAttributes := record.FindAttributes(mft.AttributeTypeData)
	if len(dataAttributes) == 0 {
		fatalf(exitCodeTechnicalError, "No $DATA attribute found in $MFT record\n")
	}

	if len(dataAttributes) > 1 {
		fatalf(exitCodeTechnicalError, "More than 1 $DATA attribute found in $MFT record\n")
	}
	
	dataAttribute := dataAttributes[0]
	if dataAttribute.Resident {
		fatalf(exitCodeTechnicalError, "Don't know how to handle resident $DATA attribute in $MFT record\n")
	}

	dataRuns, err := mft.ParseDataRuns(dataAttribute.Data)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to parse dataruns in $MFT $DATA record: %v\n", err)
	}

	if len(dataRuns) == 0 {
		fatalf(exitCodeTechnicalError, "No dataruns found in $MFT $DATA record\n")
	}

	fragments := mft.DataRunsToFragments(dataRuns, bytesPerCluster)
	totalLength := int64(0)
	for _, frag := range fragments {
		totalLength += int64(frag.Length)
	}

	out, err := createFileIfNotExist(outfile)
	if err != nil {
		fatalf(exitCodeFunctionalError, "Unable to open output file: %v\n", err)
	}
	defer out.Close()

	n, err := io.Copy(out, fragment.NewReader(in, fragments))
	if err != nil {
		fatalf(exitCodeTechnicalError, "Error copying data to output file: %v\n", err)
	}

	if n != totalLength {
		fatalf(exitCodeTechnicalError, "Expected to copy %d bytes, but copied only %d\n", totalLength, n)
	}
}

func createFileIfNotExist(outfile string) (*os.File, error) {
	return os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
}

func fatalf(exitCode int, format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(exitCode)
}
