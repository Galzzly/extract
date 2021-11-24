package extract

import (
	"fmt"
)

// Format to validate file extension.
type Format interface {
	CheckFormat(filename string) error
}

// formats lists the archive formats that the tool can work with
var formats = []Format{
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
	f, err := ByFormat(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

/*
	Will return an new instance of the file format based on the
	magic numbers found in the file.
*/
func ByFormat(filename string) (interface{}, error) {
	var ext interface{}
	for _, c := range formats {
		if err := c.CheckFormat(filename); err == nil {
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
