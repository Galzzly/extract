package extract

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/Galzzly/extract/internal/magic"
)

type Tar struct {
	MkdirAll bool

	tr            *tar.Reader
	readerWrapFn  func(io.Reader) (io.Reader, error)
	cleanupWrapFn func()
}

/*
	CheckExtension will check the file sent to the funcion
	against the magic numbers for Tar. If the file is a Tar
	the function will not return any error.
*/
func (*Tar) CheckExtension(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of filename
	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a Tar file
	var m = newMime("Tar", magic.Tar)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a .tar file", filename)
	}
	fmt.Println("This is a tar bundle", filename)
	return nil
}

/*
	Extract will extract the file sent to the function
*/
func (t *Tar) Extract(filename, destination string) (err error) {
	destination, err = t.topLevelDir(filename, destination)
	if err != nil {
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("problems opening the tar archive %s: %v", filename, err)
	}
	defer f.Close()

	err = t.Open(f)

	for {
		err = t.untarNextFile(destination)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("problem extracting file: %v", err)
		}
	}
	return nil
}

/*
	topLevelDir will evaluate contents of the Rar file and checks for
	a common root directory. If the root directory is found, the
	destination will be modified to be relative to the root directory.
*/
func (t *Tar) topLevelDir(source, destination string) (string, error) {
	f, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %v", source, err)
	}
	defer f.Close()

	// Open the Reader
	r := io.Reader(f)
	if t.readerWrapFn != nil {
		r, err = t.readerWrapFn(r)
		if err != nil {
			return "", fmt.Errorf("problem with the wrapping reader: %v", err)
		}
	}
	if t.cleanupWrapFn != nil {
		defer t.cleanupWrapFn()
	}

	tr := tar.NewReader(r)

	// Get the files in the Tar archive
	var files []string
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("issue scanning tar file listings: %v", err)
		}
		files = append(files, f.Name)
	}

	if TopLevels(files) {
		destination = filepath.Join(destination, DirFromFile(source))
	}

	return destination, nil
}

func (t *Tar) untarNextFile(destination string) (err error) {
	f, err := t.Read()
	if err != nil {
		return
	}

	h, ok := f.Header.(*tar.Header)
	if !ok {
		return fmt.Errorf("expected header to be *tar.Header but found %T", f.Header)
	}

	err = CheckPath(destination, h.Name)
	if err != nil {
		return fmt.Errorf("checking path: %v", err)
	}

	return t.untarFile(f, destination, h)
}

func (t *Tar) untarFile(f File, destination string, h *tar.Header) (err error) {
	dest := filepath.Join(destination, h.Name)

	switch h.Typeflag {
	case tar.TypeDir:
		return Mkdir(dest, f.Mode())
	case tar.TypeReg, tar.TypeRegA, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
		return WriteFile(dest, f, f.Mode())
	case tar.TypeSymlink:
		return WriteSymlink(dest, h.Linkname)
	case tar.TypeLink:
		return WriteHardlink(dest, h.Linkname)
	case tar.TypeXGlobalHeader:
		return nil
	default:
		return fmt.Errorf("%s: unknown type flag: %c", h.Name, h.Typeflag)
	}
}

func (t *Tar) Open(in io.Reader) (err error) {
	if t.tr != nil {
		return fmt.Errorf("tar archive is already open")
	}
	if t.readerWrapFn != nil {
		fmt.Println("this bit")
		in, err = t.readerWrapFn(in)
		if err != nil {
			return fmt.Errorf("issue wrapping file reader: %v", err)
		}
	}
	t.tr = tar.NewReader(in)
	return nil
}

func (t *Tar) Read() (f File, err error) {
	if t.tr == nil {
		return File{}, fmt.Errorf("tar archive is not open")
	}

	h, err := t.tr.Next()
	if err != nil {
		return File{}, err
	}

	file := File{
		FileInfo:   h.FileInfo(),
		Header:     h,
		ReadCloser: ReadFakeCloser{t.tr},
	}
	return file, nil
}

func (t *Tar) Close() {
	if t.tr == nil {
		t.tr = nil
	}
	if t.cleanupWrapFn != nil {
		t.cleanupWrapFn()
	}
}

// Not entirely sure that I need this here. May need to remove the contents
// of the struct, and just return it empty
func NewTar() *Tar {
	return &Tar{
		MkdirAll: true,
	}
}
