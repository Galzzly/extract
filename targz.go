package extract

import (
	"compress/gzip"
	"fmt"
	"strings"
)

// TarGz compresses a tar archive
type TarGz struct {
	*Tar
	CompressionLevel int
	SingleThread     bool
}

func (*TarGz) CheckExtension(filename string) error {
	if !strings.HasSuffix(filename, ".tar.gz") &&
		!strings.HasSuffix(filename, ".tgz") {
		return fmt.Errorf("%s is not a .tar.gz or .tgz file", filename)
	}
	return nil
}

func NewTarGz() *TarGz {
	return &TarGz{
		CompressionLevel: gzip.DefaultCompression,
		Tar:              NewTar(),
	}
}
