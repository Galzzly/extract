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

type Gz struct {
	CompressionLevel int
	SingleThread     bool
}

/*
	CheckExtension will check the file sent to the function
	against the magic numbers for GZip. If the file is a GZip
	the function will not return any error.
*/
func (gz *Gz) CheckExtension(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of filename
	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a GZip file
	var m = newMime("Gzip", magic.Gz)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a gzip file", filename)
	}
	fmt.Println("This is a gzip bundle", filename)
	return nil
}

/*
	Extract will extract the file sent to the function
*/
func (gz *Gz) Extract(filename, destination string) (err error) {
	// Open the file in filename. We can assume if you've got this
	// far that the file exists.
	f, _ := os.Open(filename)
	defer f.Close()

	// Open the destination file for writing
	out, err := os.Create(destination)
	if err != nil {
		return err
	}

	// Open the Reader
	r, err := pgzip.NewReader(f)
	if err != nil {
		return err
	}
	defer r.Close()

	// Write out the file
	_, err = io.Copy(out, r)
	return
}

func NewGz() *Gz {
	return &Gz{
		CompressionLevel: gzip.DefaultCompression,
	}
}
