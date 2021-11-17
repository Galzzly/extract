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

func (tgz *TarGz) CheckExtension(filename string) (err error) {
	l := atomic.LoadUint32(&readLimit)

	gh, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()
	var gm = newMime("Gzip", magic.Gz)
	if !gm.detector(gh, l) {
		return fmt.Errorf("%s is not a gzip bundle", filename)
	}
	f, _ := os.Open(filename)
	err = tgz.Open(f)
	th, err := GetHeader(f, l)
	if err != nil {
		return fmt.Errorf("problem looking at the underlying file in %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()
	var tm = newMime("Tar", magic.Tar)
	if !tm.detector(th, l) {
		return fmt.Errorf("%s is not a tar file", filename)
	}
	return nil
}

func (tgz *TarGz) Open(in io.Reader) (err error) {
	tgz.wrapReader()
	return tgz.Tar.Open(in)
}

func (tgz *TarGz) wrapReader() {
	var gzr io.ReadCloser
	tgz.Tar.readerWrapFn = func(r io.Reader) (gzr io.Reader, err error) {
		gzr, err = pgzip.NewReader(r)
		return
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
