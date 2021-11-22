package extract

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

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

// ReadFakeCloser is an io.Reader that has
// a no-op close method to satisfy the
// io.ReadCloser interface.
type ReadFakeCloser struct {
	io.Reader
}

// Close implements io.Closer.
func (rfc ReadFakeCloser) Close() error { return nil }

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

/*
	TopLevels reads a slice of paths in, and returns true if there are
	multiple top-level directories.
*/
func TopLevels(paths []string) bool {
	if len(paths) < 2 {
		return false
	}
	var top string
	for _, p := range paths {
		p = strings.TrimPrefix(strings.Replace(p, `\`, "/", -1), "/")
		for {
			next := path.Dir(p)
			if next == "." {
				break
			}
			p = next
		}
		if top == "" {
			top = p
		}
		if p != top {
			return true
		}
	}
	return false
}

func DirFromFile(filename string) (dir string) {
	dir = filepath.Base(filename)
	fd := strings.Index(dir, ".")
	if fd > -1 {
		dir = dir[:fd]
	}
	return dir
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func CheckPath(destination, filename string) (err error) {
	destination, _ = filepath.Abs(destination)
	dest := filepath.Join(destination, filename)
	if !strings.HasPrefix(dest, destination) {
		return &IllegalPathError{dest, filename}
	}
	return nil
}

func IsSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

func Mkdir(path string, mode os.FileMode) error {
	err := os.MkdirAll(path, mode)
	if err != nil {
		return fmt.Errorf("%s: error creating directory %v", path, err)
	}
	return nil
}

func WriteFile(destination string, in io.Reader, mode os.FileMode) (err error) {
	err = Mkdir(filepath.Dir(destination), 0755)
	if err != nil {
		return fmt.Errorf("error creating parent directories: %v", err)
	}
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("%s: error creating file: %v", destination, err)
	}
	defer out.Close()

	err = out.Chmod(mode)
	if err != nil {
		return fmt.Errorf("%s: error setting permissions: %v", destination, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: error writing file: %v", destination, err)
	}

	return nil
}

func WriteSymlink(destination, link string) (err error) {
	err = Mkdir(filepath.Dir(destination), 0755)
	if err != nil {
		return fmt.Errorf("%s: error creating parent directories: %v", destination, err)
	}

	_, err = os.Lstat(destination)
	if err == nil {
		err = os.Remove(destination)
		if err != nil {
			return fmt.Errorf("%s: error removing existing symlink: %v", destination, err)
		}
	}
	err = os.Symlink(link, destination)
	if err != nil {
		return fmt.Errorf("%s: error creating symlink: %v", destination, err)
	}
	return nil
}

func WriteHardlink(destination, link string) (err error) {
	err = Mkdir(filepath.Dir(destination), 0755)
	if err != nil {
		return fmt.Errorf("%s: error creating parent directories: %v", destination, err)
	}
	_, err = os.Lstat(destination)
	if err == nil {
		err = os.Remove(destination)
		if err != nil {
			return fmt.Errorf("%s: error removing existing symlink: %v", destination, err)
		}
	}
	err = os.Link(link, destination)
	if err != nil {
		return fmt.Errorf("%s: error creating symlink: %v", destination, err)
	}
	return nil
}
