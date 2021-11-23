package extract

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Galzzly/extract/internal/magic"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
	"github.com/vbauerster/mpb/v7"
)

// ZipCompressionMethod Compression type
type ZipCompressionMethod uint16

/*
	Compression methods.
	see https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
	Note LZMA: Disabled - because 7z isn't able to unpack ZIP+LZMA ZIP+LZMA2 archives made this way - and vice versa.
*/
const (
	Store   ZipCompressionMethod = 0
	Deflate ZipCompressionMethod = 8
	BZIP2   ZipCompressionMethod = 12
	LZMA    ZipCompressionMethod = 14
	ZSTD    ZipCompressionMethod = 93
	XZ      ZipCompressionMethod = 95
)

type Zip struct {
	CompressionLevel    int
	MkdirAll            bool
	SeletiveCompression bool
	FileMethod          uint16

	zr   *zip.Reader
	ridx int
}

/*
	CheckExtension will check the file sent to the funcion
	against the magic numbers for Zip. If the file is a Zip
	the function will not return any error.
*/
func (*Zip) CheckExtension(filename string) error {
	l := atomic.LoadUint32(&readLimit)

	// Get the header of filename
	h, err := GetFileHeader(filename, l)
	if err != nil {
		return fmt.Errorf("problem looking at %s", filename)
	}
	mu.Lock()
	defer mu.Unlock()

	// Check if the file is a Zip file
	var m = newMime("Zip", magic.Zip)
	if !m.detector(h, l) {
		return fmt.Errorf("%s is not a Zip file", filename)
	}
	return nil
}

func regDecomp(zr *zip.Reader) {
	zr.RegisterDecompressor(uint16(ZSTD), func(r io.Reader) io.ReadCloser {
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil
		}
		return zr.IOReadCloser()
	})
	zr.RegisterDecompressor(uint16(BZIP2), func(r io.Reader) io.ReadCloser {
		bz2, err := bzip2.NewReader(r, nil)
		if err != nil {
			return nil
		}
		return bz2
	})
	zr.RegisterDecompressor(uint16(XZ), func(r io.Reader) io.ReadCloser {
		xr, err := xz.NewReader(r)
		if err != nil {
			return nil
		}
		return ioutil.NopCloser(xr)
	})
}

/*
	Extract will extract the file sent to the function
*/
func (z *Zip) Extract(filename, destination string, p *mpb.Progress, start time.Time) (err error) {
	b := AddNewBar(p, filename, start)
	destination, err = z.topLevelDir(filename, destination)
	if err != nil {
		b.Abort(true)
		return
	}

	for {
		err = z.unzipNextFile(destination)
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Abort(true)
			return fmt.Errorf("error reading file in zip archive: %v", err)
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
func (z *Zip) topLevelDir(filename, destination string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %v", filename, err)
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("error getting file info for %s: %v", filename, err)
	}

	err = z.Open(f, fInfo.Size())
	if err != nil {
		return "", fmt.Errorf("error opening archive for reading: %v", err)
	}
	defer z.Close()

	var files = make([]string, len(z.zr.File))
	for i := range z.zr.File {
		files[i] = z.zr.File[i].Name
	}

	if TopLevels(files) {
		destination = filepath.Join(destination, DirFromFile(filename))
	}

	return destination, nil
}

/*
	unzipNextFile will read the next file in the Rar archive, check the path
	and move on to perform the extraction via unzipFile
*/
func (z *Zip) unzipNextFile(destination string) (err error) {
	f, err := z.Read()
	if err != nil {
		return
	}
	defer f.Close()

	fh, ok := f.Header.(zip.FileHeader)
	if !ok {
		return fmt.Errorf("expected header to be *zip.FileHeader but found %T", f.Header)
	}

	err = CheckPath(destination, fh.Name)
	if err != nil {
		return fmt.Errorf("checking path: %v", err)
	}

	return z.unzipFile(f, destination, &fh)
}

/*
	unzipFile will extract the file sent to the function
*/
func (z *Zip) unzipFile(f File, destination string, fh *zip.FileHeader) (err error) {
	destination = filepath.Join(destination, fh.Name)

	if f.IsDir() {
		return Mkdir(destination, fh.Mode())
	}

	if IsSymlink(fh.FileInfo()) {
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, f)
		if err != nil {
			return fmt.Errorf("%s: error reading symlink target: %v", fh.Name, err)
		}
		return WriteSymlink(destination, strings.TrimSpace(buf.String()))
	}

	return WriteFile(destination, f, fh.Mode())
}

/*
	Open will open the Zip file for reading
*/
func (z *Zip) Open(in io.Reader, size int64) (err error) {
	inRA, ok := in.(io.ReaderAt)
	if !ok {
		return fmt.Errorf("input is not a ReaderAt")
	}
	if z.zr != nil {
		return fmt.Errorf("zip archive is already open for reading")
	}

	z.zr, err = zip.NewReader(inRA, size)
	if err != nil {
		return fmt.Errorf("error creating zip reader: %v", err)
	}

	regDecomp(z.zr)
	z.ridx = 0
	return nil
}

/*
	Read will read the next file in the Zip archive
*/
func (z *Zip) Read() (f File, err error) {
	if z.zr == nil {
		return File{}, fmt.Errorf("zip archive is not open for reading")
	}

	if z.ridx >= len(z.zr.File) {
		return File{}, io.EOF
	}

	zf := z.zr.File[z.ridx]
	z.ridx++

	f = File{
		FileInfo: zf.FileInfo(),
		Header:   zf.FileHeader,
	}

	rc, err := zf.Open()
	if err != nil {
		return f, fmt.Errorf("%s: opening compressed fie: %v", f.Name(), err)
	}
	f.ReadCloser = rc
	return f, nil
}

/*
	Close will close the Zip archive
*/
func (z *Zip) Close() {
	if z.zr == nil {
		z.zr = nil
	}
}

func NewZip() *Zip {
	return &Zip{
		CompressionLevel:    flate.DefaultCompression,
		MkdirAll:            true,
		SeletiveCompression: true,
		FileMethod:          8,
	}
}
