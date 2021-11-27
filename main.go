package main

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT (see 4.3.7)
type LocalFileHeader struct {
	Signature        uint32
	Version          uint16
	BitFlag          uint16
	Compression      uint16
	LastModifiedTime uint16
	LastModifiedDate uint16
	Crc32            uint32
	CompressedSize   uint32
	UncompressedSize uint32
	FileNameLength   uint16
	ExtraFieldLength uint16
}

type FileEntry struct {
	Header   LocalFileHeader
	Filename string
	Modified time.Time
	Contents []byte
}

func msdosTimeToGoTime(d uint16, t uint16) time.Time {
	seconds := int((t & 0x1F) * 2)
	minutes := int((t >> 5) & 0x3F)
	hours := int(t >> 11)

	day := int(d & 0x1F)
	month := time.Month((d >> 5) & 0x0F)
	year := int((d>>9)&0x7F) + 1980
	return time.Date(year, month, day, hours, minutes, seconds, 0, time.Local)
}

var errCentralDirectory = errors.New("central Directory")

func parseEntry(f *os.File) (FileEntry, error) {
	entry := FileEntry{}
	// read the local file header
	err := binary.Read(f, binary.LittleEndian, &entry.Header)
	if err != nil {
		return entry, fmt.Errorf("could not read local file header length: %s", err)
	}

	if entry.Header.Signature == 0x02014b50 {
		// if we encountered the central directory, this is the end  of the file
		return entry, errCentralDirectory
	} else if entry.Header.Signature != 0x04034b50 {
		// We expected a local file header here.
		return entry, errors.New("not a zipfile")
	}

	entry.Modified = msdosTimeToGoTime(entry.Header.LastModifiedDate, entry.Header.LastModifiedTime)
	// read the filename
	filename := make([]byte, entry.Header.FileNameLength)
	_, err = f.Read(filename)
	if err != nil {
		return entry, fmt.Errorf("could not read file name: %s", err)
	}
	entry.Filename = string(filename)

	// skip the extrafield
	_, err = f.Seek(int64(entry.Header.ExtraFieldLength), 1)
	if err != nil {
		return entry, fmt.Errorf("could not skip extrafield: %s", err)
	}

	// read the (compressed?) file contents
	data := make([]byte, entry.Header.CompressedSize)
	_, err = f.Read(data)
	if err != nil {
		return entry, err
	}

	if entry.Header.Compression == 0 {
		entry.Contents = data
	} else if entry.Header.Compression == 8 {
		flateReader := flate.NewReader(bytes.NewReader(data))
		defer flateReader.Close()
		read, err := ioutil.ReadAll(flateReader)
		if err != nil {
			return entry, fmt.Errorf("error reading compressed data: %s", err)
		}
		entry.Contents = read
	} else {
		return entry, fmt.Errorf("unsupported compression method was found: %d", entry.Header.Compression)
	}

	return entry, nil
}

func printData(f *os.File, showContent bool) error {
	for {
		entry, err := parseEntry(f)
		if err == errCentralDirectory {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s\n", entry.Modified, entry.Filename)
		if showContent {
			fmt.Print(string(entry.Contents))
		}
	}
	return nil
}

func main() {
	var in *os.File
	var bShowContent = flag.Bool("c", false, "Show full file content")
	flag.Parse()
	if filename := flag.Arg(0); filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println("Could not open file: ", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	} else {
		in = os.Stdin
	}
	err := printData(in, *bShowContent)
	if err != nil {
		fmt.Println("Could not parse file: ", err)
		os.Exit(1)
	}
}
