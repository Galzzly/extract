package extract

import (
	"fmt"
	"io"
	"os"

	"github.com/dpaks/goworkers"
)

var readLimit uint32 = 3072

type Extractor interface {
	Extract(filename, dest string) error
}

type Reader interface {
	Open(in io.Reader) error
	Read() (File, error)
	Close() error
}

type File struct {
	os.FileInfo
	Header interface{}
	io.ReadCloser
}

func Extract(fileList *[]string, destDir string, numC uint32) (err error) {
	opts := goworkers.Options{Workers: numC}
	gw := goworkers.New(opts)

	go func() {
		for err := range gw.ErrChan {
			fmt.Println(err)
		}
	}()

	for _, f := range *fileList {
		file := f
		gw.SubmitCheckError(func() error {
			return extractFile(file, destDir)
		})
	}

	gw.Stop(true)

	return
}

func extractFile(file, destDir string) (err error) {
	iface, err := GetFormat(file)
	if err != nil {
		return err
	}

	u, ok := iface.(Extractor)
	if !ok {
		return fmt.Errorf("%s is not a supported file format", file)
	}
	err = u.Extract(file, destDir)
	if err != nil {
		return err
	}

	return nil
}
