package extract

import (
	"fmt"
	"strings"

	"github.com/dsnet/compress/bzip2"
)

type Bz2 struct {
	CompressionLevel int
}

func (*Bz2) CheckExtension(filename string) error {
	if !strings.HasSuffix(filename, ".bz2") {
		return fmt.Errorf("%s is not a .bz2 file", filename)
	}
	return nil
}

func NewBz2() *Bz2 {
	return &Bz2{
		CompressionLevel: bzip2.DefaultCompression,
	}
}
