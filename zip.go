package extract

import (
	"compress/flate"
	"fmt"
	"strings"
)

type Zip struct {
	CompressionLevel    int
	MkdirAll            bool
	SeletiveCompression bool
	FileMethod          uint16
}

func (*Zip) CheckExtension(filename string) error {
	if !strings.HasSuffix(filename, ".zip") {
		return fmt.Errorf("%s is not a .zip file", filename)
	}
	return nil
}

func NewZip() *Zip {
	return &Zip{
		CompressionLevel:    flate.DefaultCompression,
		MkdirAll:            true,
		SeletiveCompression: true,
		FileMethod:          8,
	}
}
