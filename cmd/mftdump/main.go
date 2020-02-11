package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	showProgress            = false
)

func main() {
	start := time.Now()
	verboseFlag := flag.Bool("v", false, "verbose; print details about what's going on")
	forceFlag := flag.Bool("f", false, "force; overwrite the output file if it already exists")
	progressFlag := flag.Bool("p", false, "progress; show progress during dumping")

	flag.Usage = printUsage
	flag.Parse()

	verbose = *verboseFlag
	overwriteOutputIfExists = *forceFlag
	showProgress = *progressFlag
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
	n, err := copy(out, fragment.NewReader(in, fragments), totalLength)
	if err != nil {
		fatalf(exitCodeTechnicalError, "Error copying data to output file: %v\n", err)
	}

	if n != totalLength {
		fatalf(exitCodeTechnicalError, "Expected to copy %d bytes, but copied only %d\n", totalLength, n)
	}
	end := time.Now()
	dur := end.Sub(start)
	printVerbose("Finished in %v\n", dur)
}

func copy(dst io.Writer, src io.Reader, totalLength int64) (written int64, err error) {
	buf := make([]byte, 1024 * 1024)
	if !showProgress {
		return io.CopyBuffer(dst, src, buf)
	}

	onePercent := float64(totalLength) / float64(100.0)
	totalSize := formatBytes(totalLength)

	// Below copied from io.copyBuffer (https://golang.org/src/io/io.go?s=12796:12856#L380)
	for {
		printProgress(written, totalSize, onePercent)

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	printProgress(written, totalSize, onePercent)
	fmt.Println()
	return written, err
}

func printProgress(n int64, totalSize string, onePercent float64) {
	percentage := float64(n) / onePercent
	barCount := int(percentage / 2.0)
	spaceCount := 50 - barCount
	fmt.Printf("\r[%s%s] %.2f%% (%s / %s)     ", strings.Repeat("|", barCount), strings.Repeat(" ", spaceCount), percentage, formatBytes(n), totalSize)
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
