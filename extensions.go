package extract

import (
	"fmt"
)

// Extension to validate file extension.
type Extension interface {
	CheckExtension(filename string) error
}

// extensions lists the extension formats that the tool can work with
var extensions = []Extension{
	&Tar{},
	&TarGz{},
	&Gz{},
	&Rar{},
	&Zip{},
	&Bz2{},
}

/*
	GetFormat returns the format of the file
*/
func GetFormat(filename string) (interface{}, error) {
	f, err := ByExtension(filename)
	if err != nil {
		return nil, err
	}

	return f, nil

}

/*
	Will return an new instance of the file format based on the
	magic numbers found in the file.
*/
func ByExtension(filename string) (interface{}, error) {
	var ext interface{}
	for _, c := range extensions {
		if err := c.CheckExtension(filename); err == nil {
			ext = c
			break
		}
	}

	switch ext.(type) {
	case *Tar:
		return NewTar(), nil
	case *TarGz:
		return NewTarGz(), nil
	case *Gz:
		return NewGz(), nil
	case *Rar:
		return NewRar(), nil
	case *Zip:
		return NewZip(), nil
	case *Bz2:
		return NewBz2(), nil
	}

	return nil, fmt.Errorf("unable to recognise format by filename: %s", filename)
}
