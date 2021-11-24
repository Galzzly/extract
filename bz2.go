package extract

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/Galzzly/extract/internal/magic"
	"github.com/dsnet/compress/bzip2"
	"github.com/vbauerster/mpb/v7"
)

type Bz2 struct {
	CompressionLevel int
}

/*
	CheckFormat will check the file sent to the function
	against the magic numbers for Bzip2. If the file is a Bzip2
	the function will not return any error.
*/
func (*Bz2) CheckFormat(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of filename
	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a BZip2 file
	var m = newMime("Bz2", magic.Bz2)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a Bzip2 file", filename)
	}
	return nil
}

/*
	Extract will extract the file sent to the function
*/
func (bz *Bz2) Extract(filename, destination string, p *mpb.Progress, start time.Time) (err error) {
	b := AddNewBar(p, filename, start)
	// Open the file in filename. We can assume if you've got this
	// far that the file exists.
	f, _ := os.Open(filename)
	defer f.Close()

	// Open the destination file for writing
	out, err := os.Create(destination)
	if err != nil {
		b.Abort(true)
		return err
	}

	// Open the Reader
	r, err := bzip2.NewReader(f, nil)
	if err != nil {
		b.Abort(true)
		return err
	}
	defer r.Close()

	// Write out the file
	_, err = io.Copy(out, r)
	if err != nil {
		b.Abort(true)
		return err
	}
	b.SetTotal(1, true)
	return
}

func NewBz2() *Bz2 {
	return &Bz2{
		CompressionLevel: bzip2.DefaultCompression,
	}
}
