package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/Galzzly/extract/v2/internal/magic"
	"github.com/nwaples/rardecode"
	"github.com/vbauerster/mpb/v7"
)

type Rar struct {
	MkdirAll bool

	rr *rardecode.Reader
	rc *rardecode.ReadCloser
}

/*
	CheckFormat will check the file sent to the function
	against the magic numbers for Rar. If the file is a Rar
	the function will not return any error.
*/
func (*Rar) CheckFormat(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of filename
	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a Rar file
	var m = newMime("Rar", magic.Rar)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a Rar file", filename)
	}
	return nil
}

/*
	Extract will extract the file sent to the function
*/
func (rar *Rar) Extract(filename, destination string, p *mpb.Progress, start time.Time) (err error) {
	b := AddNewBar(p, filename, start)
	// Check for a common root, and return a modified destination
	// so that we don't clobber the destination directory
	destination, err = rar.topLevelDir(filename, destination)
	if err != nil {
		b.Abort(true)
		return
	}

	// Open up the Rar file for reading
	// Supporting multi-volume archives.
	err = rar.OpenRarFile(filename)
	if err != nil {
		b.Abort(true)
		return fmt.Errorf("unable to open rar file for reading: %v", err)
	}
	defer rar.Close()

	for {
		err = rar.unrarNextFile(destination)
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Abort(true)
			return fmt.Errorf("issue reading file in rar archive: %v", err)
		}
	}
	b.SetTotal(1, true)
	return nil
}

/*
	topLevelDir will evaluate contents of the Rar file and checks for
	a common root directory. If the root directory is found, the
	destination will be modified to be relative to the root directory.
*/
func (rar *Rar) topLevelDir(source, destination string) (string, error) {
	f, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %v", source, err)
	}
	defer f.Close()

	r, err := rardecode.NewReader(f, "")
	if err != nil {
		return "", fmt.Errorf("unable to open rar archive: %v", err)
	}

	// Get the files in the Rar file
	var files []string
	for {
		f, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("issue scanning rar file listings: %v", err)
		}
		files = append(files, f.Name)
	}

	if TopLevels(files) {
		destination = filepath.Join(destination, DirFromFile(source))
	}

	return destination, nil
}

/*
	unrarNextFile will read the next file in the Rar archive, check the path
	and move on to perform the extraction via unrarFile
*/
func (rar *Rar) unrarNextFile(destination string) (err error) {
	f, err := rar.Read()
	if err != nil {
		return
	}
	defer f.Close()

	fh, ok := f.Header.(*rardecode.FileHeader)
	if !ok {
		return fmt.Errorf("expected header to be *rardecode.FileHeader but found %T", f.Header)
	}

	// err = rar.CheckPath(destination, fh.Name)
	err = CheckPath(destination, fh.Name)
	if err != nil {
		return fmt.Errorf("checking path: %v", err)
	}

	return rar.unrarFile(f, filepath.Join(destination, fh.Name))
}

/*
	unrarFile will extract the file sent to the function
*/
func (rar *Rar) unrarFile(f File, destination string) (err error) {
	fh, ok := f.Header.(*rardecode.FileHeader)
	if !ok {
		return fmt.Errorf("expected header to be *rardecode.FileHeader but found %T", f.Header)
	}

	if f.IsDir() {
		return Mkdir(destination, fh.Mode())
	}

	if (fh.Mode() & os.ModeSymlink) != 0 {
		return nil
	}

	return WriteFile(destination, rar.rr, fh.Mode())
}

/*
	OpenRarFile will open the Rar file for reading
*/
func (rar *Rar) OpenRarFile(file string) (err error) {
	if rar.rr != nil {
		return fmt.Errorf("rar archive is already open for reading")
	}

	rar.rc, err = rardecode.OpenReader(file, "")
	if err != nil {
		return
	}
	rar.rr = &rar.rc.Reader
	return nil
}

/*
	Read will read the next file in the Rar archive
*/
func (rar *Rar) Read() (f File, err error) {
	if rar.rr == nil {
		return File{}, fmt.Errorf("rar archive is not open for reading")
	}

	fh, err := rar.rr.Next()
	if err != nil {
		return File{}, err
	}

	f = File{
		FileInfo:   rarInfo{fh},
		Header:     fh,
		ReadCloser: ReadFakeCloser{rar.rr},
	}
	return f, nil
}

/*
	Close will close the Rar archive
*/
func (rar *Rar) Close() (err error) {
	if rar.rc != nil {
		rc := rar.rc
		rar.rc = nil
		err = rc.Close()
	}
	if rar.rr != nil {
		rar.rr = nil
	}
	return
}

func NewRar() *Rar {
	return &Rar{
		MkdirAll: true,
	}
}

type rarInfo struct {
	fh *rardecode.FileHeader
}

func (ri rarInfo) Name() string       { return ri.fh.Name }
func (ri rarInfo) Size() int64        { return ri.fh.UnPackedSize }
func (ri rarInfo) Mode() os.FileMode  { return ri.fh.Mode() }
func (ri rarInfo) ModTime() time.Time { return ri.fh.ModificationTime }
func (ri rarInfo) IsDir() bool        { return ri.fh.IsDir }
func (ri rarInfo) Sys() interface{}   { return nil }
