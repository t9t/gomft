package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/t9t/gomft/bootsect"
	"github.com/t9t/gomft/fragment"
	"github.com/t9t/gomft/mft"
)

const supportedOemId = "NTFS    "

const (
	exitCodeUserError int = iota + 2
	exitCodeFunctionalError
	exitCodeTechnicalError
)

const isWin = runtime.GOOS == "windows"

var (
	// flags
	verbose                 = false
	overwriteOutputIfExists = false
)

func main() {
	verboseFlag := flag.Bool("v", false, "verbose; print details about what's going on")
	forceFlag := flag.Bool("f", false, "force; overwrite the output file if it already exists")

	flag.Usage = printUsage
	flag.Parse()

	verbose = *verboseFlag
	overwriteOutputIfExists = *forceFlag
	args := flag.Args()

	if len(args) != 2 {
		printUsage()
		os.Exit(exitCodeUserError)
		return
	}

	volume := args[0]
	if isWin {
		volume = `\\.\` + volume
	}
	outfile := args[1]

	in, err := os.Open(volume)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to open volume using path %s: %v\n", volume, err)
	}
	defer in.Close()

	printVerbose("Reading boot sector\n")
	bootSectorData := make([]byte, 512)
	_, err = io.ReadFull(in, bootSectorData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to read boot sector: %v\n", err)
	}

	printVerbose("Read %d bytes of boot sector, parsing boot sector\n", len(bootSectorData))
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

	mftSizeInBytes := bootSector.FileRecordSegmentSize.ToBytes(bytesPerCluster)
	printVerbose("Reading $MFT file record at position %d (size: %d bytes)\n", mftPosInBytes, mftSizeInBytes)
	mftData := make([]byte, mftSizeInBytes)
	_, err = io.ReadFull(in, mftData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to read $MFT record: %v\n", err)
	}

	printVerbose("Parsing $MFT file record\n")
	record, err := mft.ParseRecord(mftData)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Unable to parse $MFT record: %v\n", err)
	}

	printVerbose("Reading $DATA attribute in $MFT file record\n")
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

	out, err := openOutputFile(outfile)
	if err != nil {
		fatalf(exitCodeFunctionalError, "Unable to open output file: %v\n", err)
	}
	defer out.Close()

	printVerbose("Copying %d bytes (%s) of data to %s\n", totalLength, formatBytes(totalLength), outfile)
	n, err := io.Copy(out, fragment.NewReader(in, fragments))
	if err != nil {
		fatalf(exitCodeTechnicalError, "Error copying data to output file: %v\n", err)
	}

	if n != totalLength {
		fatalf(exitCodeTechnicalError, "Expected to copy %d bytes, but copied only %d\n", totalLength, n)
	}
	printVerbose("Finished\n")
}

func openOutputFile(outfile string) (*os.File, error) {
	if overwriteOutputIfExists {
		return os.Create(outfile)
	} else {
		return os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	}
}

func printUsage() {
	out := os.Stderr
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(out, "\nusage: %s [flags] <volume> <output file>\n\n", exe)
	fmt.Fprintln(out, "Dump the MFT of a volume to a file. The volume should be NTFS formatted.")
	fmt.Fprintln(out, "\nFlags:")

	flag.PrintDefaults()

	fmt.Fprintf(out, "\nFor example: ")
	if isWin {
		fmt.Fprintf(out, "%s -v -f C: D:\\c.mft\n", exe)
	} else {
		fmt.Fprintf(out, "%s -v -f /dev/sdb1 ~/sdb1.mft\n", exe)
	}
}

func fatalf(exitCode int, format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(exitCode)
}

func printVerbose(format string, v ...interface{}) {
	if verbose {
		fmt.Printf(format, v...)
	}
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	if b < 1048576 {
		return fmt.Sprintf("%.2fKiB", float32(b)/float32(1024))
	}
	if b < 1073741824 {
		return fmt.Sprintf("%.2fMiB", float32(b)/float32(1048576))
	}
	return fmt.Sprintf("%.2fGiB", float32(b)/float32(1073741824))
}
