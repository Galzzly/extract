package extract

import (
	"archive/tar"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/Galzzly/extract/internal/magic"
)

type Tar struct {
	MkdirAll bool

	tr            *tar.Reader
	readerWrapFn  func(io.Reader) (io.Reader, error)
	cleanupWrapFn func()
}

func (*Tar) CheckExtension(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()
	var m = newMime("Tar", magic.Tar)
	fmt.Println(m.detector(h, l))
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a .tar file", filename)
	}
	fmt.Println("This is a tar bundle", filename)
	return nil
}

func (t *Tar) Extract(filename, dest string) error {

	return nil
}

func (t *Tar) Open(in io.Reader) (err error) {
	if t.tr != nil {
		return fmt.Errorf("tar archive is already open")
	}

	if t.readerWrapFn != nil {
		in, err = t.readerWrapFn(in)
		if err != nil {
			return fmt.Errorf("issue wrapping file reader: %v", err)
		}
	}
	t.tr = tar.NewReader(in)
	return nil
}

// Not entirely sure that I need this here. May need to remove the contents
// of the struct, and just return it empty
func NewTar() *Tar {
	return &Tar{
		MkdirAll: true,
	}
}
