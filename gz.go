package extract

import (
	"compress/gzip"
	"fmt"
	"sync/atomic"

	"github.com/Galzzly/extract/internal/magic"
)

type Gz struct {
	CompressionLevel int
	SingleThread     bool
}

func (gz *Gz) CheckExtension(filename string) error {
	// if filepath.Ext(filename) != ".gz" {
	// 	return fmt.Errorf("%s is not a .gz file", filename)
	// }
	l := atomic.LoadUint32(&readLimit)

	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()
	var m = newMime("Gzip", magic.Gz)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a gzip file", filename)
	}
	fmt.Println("This is a gzip bundle", filename)
	return nil
}

func NewGz() *Gz {
	return &Gz{
		CompressionLevel: gzip.DefaultCompression,
	}
}
