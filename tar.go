package extract

import (
	"fmt"
	"strings"
)

type Tar struct {
	MkdirAll bool
}

func (*Tar) CheckExtension(filename string) error {
	if !strings.HasSuffix(filename, ".tar") {
		return fmt.Errorf("%s is not a .tar file", filename)
	}
	return nil
}

func (t *Tar) Extract(filename, dest string) error {

	return nil
}

// Not entirely sure that I need this here. May need to remove the contents
// of the struct, and just return it empty
func NewTar() *Tar {
	return &Tar{
		MkdirAll: true,
	}
}
