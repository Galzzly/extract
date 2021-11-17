// package extract

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dsnet/compress/bzip2"
	"github.com/h2non/filetype"
	"github.com/nwaples/rardecode"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

var formats = []string{"application/gzip",
	"application/x-tar",
	"application/zip",
	"application/x-rar",
	"application/vnd.rar",
	"application/x-bzip2",
}

type WriteCounter struct {
	n   int
	bar *mpb.Bar
}

func (wc *WriteCounter) Write(p []byte) (n int, err error) {
	wc.n += len(p)
	wc.bar.IncrBy(len(p))
	return wc.n, nil
}

/*
Central place to keep all of the extraction functions to be called on from main.go
	gzip - will need to loop around again just in case there is a tar bundle within
	tar
	zip
	7za
	rar
	bzip
*/

/*
	Start off with the main function, that will process.
*/

func Extract(fileList []string, destDir string, numC int) (err error) {
	var wg sync.WaitGroup
	wg.Add(len(fileList))

	p := mpb.New(mpb.WithWaitGroup(&wg))

	workers := make(chan int, numC)

	for _, f := range fileList {
		go extract(f, destDir, &wg, p, workers)
	}

	p.Wait()
	wg.Wait()
	close(workers)

	return
}

func extract(source, destDir string, wg *sync.WaitGroup, p *mpb.Progress, worker chan int) {
	defer wg.Done()
	worker <- 1
	buf, _ := ioutil.ReadFile(source)
	format, _ := filetype.Match(buf)
	if filetype.IsArchive(buf) && find(format.MIME.Value) {
		_ = extr(source, destDir, format.Extension, format.MIME.Value, p)
	}

	<-worker
}

func find(frmt string) bool {
	for _, f := range formats {
		if f == frmt {
			return true
		}
	}
	return false
}

func extr(source, destDir, ext, frmt string, p *mpb.Progress) (err error) {
	start := time.Now()
	b := p.AddBar(
		int64(100),
		mpb.BarClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(source+":", decor.WC{W: len(source) + 2, C: decor.DidentRight}),
			decor.OnComplete(decor.Name("Extracting", decor.WCSyncSpaceR), fmt.Sprintf("Done! (%s)", time.Since(start))),
			decor.Percentage(decor.WCSyncSpace),
		),
	)

	counter := &WriteCounter{bar: b}

	switch frmt {
	// Gzip files
	case "application/gzip":
		f, err := os.Open(source)
		if err != nil {
			return err
		}
		defer f.Close()
		err = Gzip(f, destDir, counter)
		if err != nil {
			return err
		}

	// Tar files
	case "application/x-tar":
		f, err := os.Open(source)
		if err != nil {
			return err
		}
		defer f.Close()
		err = Tar(f, destDir, counter)
		if err != nil {
			return err
		}
	// Zip files
	case "application/zip":
		err = Zip(source, destDir, counter)
	// Rar files, without password
	case "application/x-rar", "application/vnd.rar":
		f, err := os.Open(source)
		if err != nil {
			return err
		}
		defer f.Close()
		err = Rar(f, destDir, counter)
		if err != nil {
			return err
		}
	// Bzip files
	case "application/x-bzip2":
		f, err := os.Open(source)
		if err != nil {
			return err
		}
		defer f.Close()
		err = Bzip(f, GetFileName(source), destDir, counter)
		if err != nil {
			return err
		}
	// Anything else, we do not process right now
	default:
		err = errors.New("unable to process right now")
		b.Completed()
		if err != nil {
			return err
		}

	}

	return nil
}

func validPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

// Extraction of Gzip files
func Gzip(src io.Reader, dst string, c *WriteCounter) error {
	zr, e := gzip.NewReader(src)
	Check(e)
	defer zr.Close()

	tr := tar.NewReader(zr)
	// Check if there is a TAR file within
	_, e = tr.Next()
	if e != nil {
		// No TAR bundle, just extract the gz

		target := filepath.Join(dst, zr.Name)

		// Check name against the path
		if !validPath(zr.Name) {
			return fmt.Errorf("bundle contained invalid name error %q", target)
		}

		opWriter, e := os.Create(target)
		if e != nil {
			return e
		}
		defer opWriter.Close()

		_, e = io.Copy(opWriter, io.TeeReader(zr, c))
		if e != nil {
			return e
		}
	} else {
		// There is a TAR bundle, work through that
		e := Tar(zr, dst, c)
		if e != nil {
			return e
		}
	}

	return nil
}

func Tar(src io.Reader, dst string, c *WriteCounter) error {
	tr := tar.NewReader(src)
	/*
		Need to track the directories with their modified times to correct
		after extraction. Otherwise, modified will update with each file
		extracted within.
	*/
	dirTimes := make(map[string]map[string]time.Time)

	for {
		header, e := tr.Next()
		if e == io.EOF {
			break // Reached the end of the archive
		}
		if e != nil {
			return e
		}

		target := filepath.Join(dst, header.Name)

		// Check the path
		if !validPath(header.Name) {
			fmt.Errorf("bundle contained invalid name error %q", target)
		}

		switch header.Typeflag {
		// If it's a directory....
		case tar.TypeDir:
			if _, e := os.Stat(target); e != nil {
				if e := os.MkdirAll(target, 0755); e != nil {
					return e
				}
				dirTimes[target] = map[string]time.Time{
					"mTime": header.ModTime,
					"aTime": header.AccessTime,
				}
			}
		// If it's a file...
		case tar.TypeReg:
			// Need to make sure that the parent directory is there.
			parentDir, _ := filepath.Split(target)
			if _, e := os.Stat(parentDir); e != nil {
				if e := os.MkdirAll(parentDir, 0755); e != nil {
					return e
				}
			}
			ftw, e := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if e != nil {
				return e
			}
			if _, e := io.Copy(ftw, io.TeeReader(tr, c)); e != nil {
				return e
			}
			ftw.Close()
			// Set the modification time & access time
			e = os.Chtimes(target, header.AccessTime, header.ModTime)
			if e != nil {
				return e
			}
		}
	}

	/*
		Now that the loop is finished, and the individual files will have the times correct
		Set the directories using the map made
	*/
	for dir := range dirTimes {
		e := os.Chtimes(dir, dirTimes[dir]["aTime"], dirTimes[dir]["mTime"])
		if e != nil {
			return e
		}
	}

	return nil
}

func Zip(src, dst string, c *WriteCounter) error {
	zp, e := zip.OpenReader(src)
	if e != nil {
		return e
	}
	defer zp.Close()

	// Cycle through the contents, with enough permissions to write the files
	for _, f := range zp.File {
		rc, e := f.Open()
		if e != nil {
			return e
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		target := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Errorf("bundle contained invalid name error %q", target)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, os.FileMode(0777))
		} else {
			os.MkdirAll(filepath.Dir(target), os.FileMode(0777))
			ftw, e := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0777))
			if e != nil {
				return e
			}
			defer func() {
				err := ftw.Close()
				if err != nil {
					return
				}
			}()

			_, e = io.Copy(ftw, io.TeeReader(rc, c))
			if e != nil {
				return e
			}
		}

		// Correct the permissions
		e = os.Chmod(target, f.Mode())
		if e != nil {
			return e
		}

		// Set the modification time & access time
		e = os.Chtimes(target, f.Modified, f.Modified)
		if e != nil {
			return e
		}
	}

	return nil
}

func Szip(src io.Reader, dst string) error {
	return nil
}

func Rar(src io.Reader, dst string, c *WriteCounter) error {
	rr, e := rardecode.NewReader(src, "")
	if e != nil {
		return e
	}

	for {
		header, e := rr.Next()
		if e == io.EOF {
			break // Reaced the end of the archive
		}
		if e != nil {
			return e
		}

		target := filepath.Join(dst, header.Name)

		// Check the path
		if !validPath(header.Name) {
			fmt.Errorf("bundle contained invalid name error %q", target)
		}

		// Check if it's a directory
		if header.IsDir {
			if _, e := os.Stat(target); e != nil {
				if e := os.MkdirAll(target, 0755); e != nil {
					return e
				}
			}
		} else {
			ftw, e := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Attributes))
			if e != nil {
				return e
			}
			if _, e := io.Copy(ftw, io.TeeReader(rr, c)); e != nil {
				return e
			}
			ftw.Close()
		}
	}

	return nil
}

func Bzip(src io.Reader, f, dst string, c *WriteCounter) error {
	br, e := bzip2.NewReader(src, nil)
	if e != nil {
		return e
	}
	defer br.Close()

	target := filepath.Join(dst, f)

	if !validPath(target) {
		fmt.Errorf("bundle contained invalid name error %q", target)
	}

	opWrite, e := os.Create(target)
	if e != nil {
		return e
	}
	defer opWrite.Close()

	_, e = io.Copy(opWrite, io.TeeReader(br, c))
	if e != nil {
		return e
	}

	return nil
}
