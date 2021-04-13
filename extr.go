package extract

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dsnet/compress/bzip2"
	"github.com/nwaples/rardecode"
)

/*
Central place to keep all of the extraction functions to be called on from main.go
	gzip - will need to loop around again just in case there is a tar bundle within
	tar
	zip
	7za
	rar
	bzip
*/

func validPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

// Extraction of Gzip files
func Gzip(src io.Reader, dst string) error {
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
			return fmt.Errorf("Bundle contained invalid name error %q", target)
		}

		opWriter, e := os.Create(target)
		Check(e)
		defer opWriter.Close()

		_, e = io.Copy(opWriter, zr)
		Check(e)
	} else {
		// There is a TAR bundle, work through that
		e := Tar(zr, dst)
		Check(e)
	}

	return nil
}

func Tar(src io.Reader, dst string) error {
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
		Check(e)

		target := filepath.Join(dst, header.Name)

		// Check the path
		if !validPath(header.Name) {
			fmt.Errorf("Bundle contained invalid name error %q", target)
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
			Check(e)
			if _, e := io.Copy(ftw, tr); e != nil {
				return e
			}
			ftw.Close()
			// Set the modification time & access time
			e = os.Chtimes(target, header.AccessTime, header.ModTime)
			Check(e)
		}
	}

	/*
		Now that the loop is finished, and the individual files will have the times correct
		Set the directories using the map made
	*/
	for dir := range dirTimes {
		e := os.Chtimes(dir, dirTimes[dir]["aTime"], dirTimes[dir]["mTime"])
		Check(e)
	}

	return nil
}

func Zip(src, dst string) error {
	zp, e := zip.OpenReader(src)
	Check(e)
	defer zp.Close()

	// Cycle through the contents, with enough permissions to write the files
	for _, f := range zp.File {
		rc, e := f.Open()
		Check(e)
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		target := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Errorf("Bundle contained invalid name error %q", target)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, os.FileMode(0777))
		} else {
			os.MkdirAll(filepath.Dir(target), os.FileMode(0777))
			ftw, e := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0777))
			Check(e)
			defer func() {
				if err := ftw.Close(); err != nil {
					panic(err)
				}
			}()

			_, e = io.Copy(ftw, rc)
			Check(e)
		}
	}

	// Cycle through again to set the permissions correctly
	for _, f := range zp.File {
		target := filepath.Join(dst, f.Name)
		e := os.Chmod(target, f.Mode())
		Check(e)
	}
	return nil
}

func Szip(src io.Reader, dst string) error {
	return nil
}

func Rar(src io.Reader, dst string) error {
	rr, e := rardecode.NewReader(src, "")
	Check(e)

	for {
		header, e := rr.Next()
		if e == io.EOF {
			break // Reaced the end of the archive
		}
		Check(e)

		target := filepath.Join(dst, header.Name)

		// Check the path
		if !validPath(header.Name) {
			fmt.Errorf("Bundle contained invalid name error %q", target)
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
			Check(e)
			if _, e := io.Copy(ftw, rr); e != nil {
				return e
			}
			ftw.Close()
		}
	}

	return nil
}

func Bzip(src io.Reader, f, dst string) error {
	br, e := bzip2.NewReader(src, nil)
	Check(e)
	defer br.Close()

	target := filepath.Join(dst, f)

	if !validPath(target) {
		fmt.Errorf("Bundle contained invalid name error %q", target)
	}

	opWrite, e := os.Create(target)
	Check(e)
	defer opWrite.Close()

	_, e = io.Copy(opWrite, br)
	Check(e)

	return nil
}
