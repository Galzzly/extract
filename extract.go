package extract

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v7"
)

var readLimit uint32 = 3072

type Extractor interface {
	Extract(filename, dest string, p *mpb.Progress, start time.Time) error
}

type File struct {
	os.FileInfo
	Header interface{}
	io.ReadCloser
}

/*
	ReadFakeCloser is an io.Reader that has
	a no-op close method to satisfy the
	io.ReadCloser interface.
*/
type ReadFakeCloser struct {
	io.Reader
}

// Close implements io.Closer.
func (rfc ReadFakeCloser) Close() error { return nil }

func Extract(fileList *[]string, destDir string, numC uint32) (err error) {
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(*fileList))

	p := mpb.New(mpb.WithWaitGroup(&wg))
	workers := make(chan int, numC)
	for _, f := range *fileList {
		file := f
		go extractFile(file, destDir, &wg, p, workers)
	}

	p.Wait()
	wg.Wait()
	close(workers)

	fmt.Println("\nExtraction complete in", time.Since(start))
	return
}

func extractFile(file, destDir string, wg *sync.WaitGroup, p *mpb.Progress, worker chan int) (err error) {
	defer wg.Done()
	worker <- 1
	start := time.Now()
	iface, err := GetFormat(file)
	if err != nil {
		<-worker
		return nil
	}

	u, _ := iface.(Extractor)

	err = u.Extract(file, destDir, p, start)
	if err != nil {
		<-worker
		fmt.Println("\r", file, "failed to extract in", time.Since(start))
		return err
	}
	fmt.Println(file, "extracted to", destDir, "in", time.Since(start))
	<-worker
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

/*
	DirFromFile returns a directory based on the filename
	provided to the function.
*/
func DirFromFile(filename string) (dir string) {
	dir = filepath.Base(filename)
	fd := strings.Index(dir, ".")
	if fd > -1 {
		dir = dir[:fd]
	}
	return dir
}

/*
	FileExists will check for the existence of a file
	provided to the function as filename
*/
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

/*
	CheckPath confirms that the path provided has not been made to
	perform a traversal attack.
*/
func CheckPath(destination, filename string) (err error) {
	destination, _ = filepath.Abs(destination)
	dest := filepath.Join(destination, filename)
	if !strings.HasPrefix(dest, destination) {
		return &IllegalPathError{dest, filename}
	}
	return nil
}

/*
	IsSymlink returns true if the file is a symlink.
*/
func IsSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

/*
	Mkdir creates a directory at the destination path, along with any
	parent directories
*/
func Mkdir(path string, mode os.FileMode) error {
	err := os.MkdirAll(path, mode)
	if err != nil {
		return fmt.Errorf("%s: error creating directory %v", path, err)
	}
	return nil
}

/*
	WriteFile writes a file to the destination path.
*/
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
	out.Close()

	return nil
}

/*
	WriteSymlink creates the symbolic link at the destination location
*/
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

/*
	WriteHardlink creates the hard link at the destination location.
*/
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
