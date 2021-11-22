package extract

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/Galzzly/extract/internal/magic"
	"github.com/klauspost/pgzip"
)

// TarGz compresses a tar archive
type TarGz struct {
	*Tar
	CompressionLevel int
	SingleThread     bool
}

/*
	CheckExtension will check the file sent to the function
	against magic numbers for Gzip & Tar. If the file is a TarGz
	the function will not return any error.
	First will check the file contains the relevant magic number
	for a GZip file, and if so, will check that the file within
	contains the magic number for a Tar file.
*/
func (tgz *TarGz) CheckExtension(filename string) (err error) {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of the GZip file
	gh, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a GZip file
	var gm = newMime("Gzip", magic.Gz)
	if !gm.detector(gh, l) {
		return fmt.Errorf("%s is not a gzip bundle", filename)
	}

	// Open the Gzip file for reading
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("problem opening %s", filename)
	}
	defer f.Close()

	// Create the pgzip reader
	r, _ := pgzip.NewReader(f)
	defer r.Close()
	if err != nil {
		return fmt.Errorf("problem opening %s", filename)
	}
	// Using the pgzip reader, get the header of the tar file
	th, err := GetHeader(r, l)
	if err != nil {
		return fmt.Errorf("problem looking at the underlying file in %s", filename)
	}

	// Check if the file within the GZip file is a tar file
	var tm = newMime("Tar", magic.Tar)
	if !tm.detector(th, l) {
		return fmt.Errorf("%s is not a tar file", filename)
	}

	fmt.Println("This is a TarGz file", filename)
	return nil
}

/*
	Extract will extract the file sent to the function
*/
func (tgz *TarGz) Extract(source, destination string) (err error) {
	tgz.wrapReader()
	return tgz.Tar.Extract(source, destination)
}

func (tgz *TarGz) Open(in io.Reader) (err error) {
	fmt.Println("I get here")
	tgz.wrapReader()
	return tgz.Tar.Open(in)
}

func (tgz *TarGz) wrapReader() {
	var gzr io.ReadCloser
	// fmt.Println("And also here")
	tgz.Tar.readerWrapFn = func(r io.Reader) (io.Reader, error) {
		//return pgzip.NewReader(r)
		gzr, err := pgzip.NewReader(r)
		return gzr, err
	}
	tgz.Tar.cleanupWrapFn = func() {
		gzr.Close()
	}
}

func NewTarGz() *TarGz {
	return &TarGz{
		CompressionLevel: gzip.DefaultCompression,
		Tar:              NewTar(),
	}
}
