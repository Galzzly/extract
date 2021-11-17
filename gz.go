package extract

import (
	"compress/gzip"
	"fmt"
	"path/filepath"
)

type Gz struct {
	CompressionLevel int
	SingleThread     bool
}

func (gz *Gz) CheckExtension(filename string) error {
	if filepath.Ext(filename) != ".gz" {
		return fmt.Errorf("%s is not a .gz file", filename)
	}
	return nil
}

func NewGz() *Gz {
	return &Gz{
		CompressionLevel: gzip.DefaultCompression,
	}
}
