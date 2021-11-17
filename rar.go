package extract

import (
	"fmt"
	"strings"
)

type Rar struct {
	MkdirAll bool
}

func (*Rar) CheckExtension(filename string) error {
	if !strings.HasSuffix(filename, ".rar") {
		return fmt.Errorf("%s is not a .rar file", filename)
	}
	return nil
}

func NewRar() *Rar {
	return &Rar{
		MkdirAll: true,
	}
}
