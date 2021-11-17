package extract

import (
	"fmt"

	"github.com/dpaks/goworkers"
)

type Extractor interface {
	Extract(filename, dest string) error
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
