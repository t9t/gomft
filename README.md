# gomft [![Build Status](https://travis-ci.com/t9t/gomft.svg?branch=master)](https://travis-ci.com/t9t/gomft) [![GoDoc](https://godoc.org/github.com/t9t/gomft?status.svg)](https://godoc.org/github.com/t9t/gomft)

gomft is Go library to parse the Master File Table (MFT) of NFTS volumes. `mftdump` is a utility to dump the MFT of a
mounted volume to a file.

Example usage reading MFT records from a file that was previously dumped with a record size of 1KB:

```go
package main

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/t9t/gomft/mft"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("Unable to open file", err)
	}
	defer f.Close()

	recordSize := 1024
	for {
		buf := make([]byte, recordSize)
		_, err := io.ReadFull(f, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalln("Unable to read record data", err)
		}

		record, err := mft.ParseRecord(buf)
		if err != nil {
			log.Fatalln("Unable to parse MFT record", err)
		}

		log.Println("Read MFT record", record.FileReference)
	}
}
```

See also: https://godoc.org/github.com/t9t/gomft/mft

## Reading from a raw volume
To read from a raw volume, you have to be root (*nix) or Administrator (Windows). In *nix you can just use the
partition device file name (eg. `/dev/sda1`) while in Windows you have to use an UNC path such as `\\.\C:`. All the
rest is the same as accessing a file (ie. `os.Open(...)`).

Note that on Windows you can only read data in multiples of the sector size, so if the sector size is 512 bytes (which
is most common), you can read 512, 1024, 1536, etc bytes at a time but not 768 for instance. Keep this in mind when
using a buffered reader, making sure the buffer size is a multiple of the sector size.

## Reading the boot sector
To read the boot sector (also known as VBR, Volume Boot Record, or $Boot file) of a volume you can use the `bootsect`
package:

```go
package main

import (
	"io"
	"log"
	"os"

	"github.com/t9t/gomft/bootsect"
)

func main() {
	f, err := os.Open(`\\.\C:`)
	if err != nil {
		log.Fatalln("Unable to open C:", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		log.Fatalln("Unable to read bootsector data", err)
	}

	bootSector, err := bootsect.Parse(buf)
	if err != nil {
		log.Fatalln("Unable to parse boot sector")
	}

	log.Printf("Boot sector of C:\\:\n%+v\n", bootSector)
}
```

See: https://godoc.org/github.com/t9t/gomft/bootsect

## Additional utilities

### Fragment reader
Use the `fragment` package to read fragmented data, for example as obtained from DataRuns in MFT records. Use
[`mft.DataRunsToFragments()`](https://godoc.org/github.com/t9t/gomft/mft#DataRunsToFragments) to translate DataRuns
into fragments.

See: https://godoc.org/github.com/t9t/gomft/fragment

### bintuil & BinReader
The `binutil` package contains some functions to help using binary data, primarily `binutil.Duplicate()` to duplicate 
a slice of bytes and `BinReader` to interpret binary data according to a certain byte order (little/big endian).

See: https://godoc.org/github.com/t9t/gomft/binutil

### utf16
The `utf16` package contains the `DecodeString` function to decode a byte slice to a string using a certain byte order.

See: https://godoc.org/github.com/t9t/gomft/utf16

## Disclaimer
This package is far from complete and the implementation scrambled together from various bits of (often conflicting)
information strewn about the internet.

**Use at your own risk!** Accessing your raw volumes could damage your data beyond repair if you are not careful! It's
probably best to dump your MFT to a file and experiment with that rather than reading your raw volumes directly.

# mftdump
The mftdump utility can be used to dump the MFT of a raw volume to a file. Download it in [the releases section](/releases).

Usage:

```
usage: mftdump [flags] <volume> <output file>

Dump the MFT of a volume to a file. The volume should be NTFS formatted.

Flags:
  -f    force; overwrite the output file if it already exists
  -p    progress; show progress during dumping
  -v    verbose; print details about what's going on

For example: mftdump -v -f /dev/sdb1 ~/sdb1.mft
```

On Windows, use it like this: `mftdump.exe -v -f C: D:\c.mft`

# References
In no particular order, these pages and programs have helped me build gomft.

- https://en.wikipedia.org/wiki/NTFS#Master_File_Table
- http://www.kcall.co.uk/ntfs/index.html
- https://flatcap.org/linux-ntfs/ntfs/index.html
- http://www.cse.scu.edu/~tschwarz/coen252_07Fall/Lectures/NTFS.html
- https://www.autoitscript.com/forum/topic/94269-mft-access-reading-parsing-the-master-file-table-on-ntfs-filesystems/
- https://www.andreafortuna.org/2017/07/18/how-to-extract-data-and-timeline-from-master-file-table-on-ntfs-filesystem/
- http://ftp.kolibrios.org/users/Asper/docs/NTFS/ntfsdoc.html
- https://docs.microsoft.com/en-us/windows/win32/fileio/master-file-table
- https://flylib.com/books/en/2.48.1/ntfs_concepts.html
- "A Journey into NTFS"
    - Part 1: https://medium.com/@bromiley/a-journey-into-ntfs-part-1-e2ac6a6367ec
    - Part 2: https://medium.com/@bromiley/ntfs-series-2b3b91faaf21
    - Part 3: https://medium.com/@bromiley/a-journey-into-ntfs-part-3-5e197a0cab58
    - Part 4: https://medium.com/@bromiley/a-journey-into-ntfs-part-4-f2865c39ac83
    - Part 5: https://medium.com/@bromiley/ntfs-part-5-13e20588af59
    - Part 6: https://medium.com/@bromiley/ntfs-part-6-43a50fad89f3
    - Part 7: https://medium.com/@bromiley/ntfs-part-7-an-ntfs-story-caf42565855b
- https://github.com/dkovar/analyzeMFT
- https://github.com/jschicht/Mft2Csv
- https://github.com/libyal/libfsntfs/blob/master/documentation/New%20Technologies%20File%20System%20(NTFS).asciidoc
