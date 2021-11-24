package extract

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

/*
	GetFileCType is a simplistic way of pulling out the MIMI type of a file.
	Does not always pull out the type, and will instead replace with
	application/octet-stream. Particularly when the filie is smaller than 512 bytes.
	As such, this should not necessarily be relied upon.
*/
func GetFileCType(out *os.File) (string, error) {
	buf := make([]byte, 512)
	_, e := out.Read(buf)
	if e != nil {
		return "", e
	}

	cType := http.DetectContentType(buf)

	return cType, nil
}

/*
	GetFileName will strip the the file name to extract to by removing the suffix.
	e.g. filename.tar becomes filename
*/
func GetFileName(f string) string {
	// Get the source file without the leading path
	_, fName := filepath.Split(f)

	// Get the suffix of the string
	s := strings.Split(fName, ".")
	suffix := "." + s[len(s)-1]

	// Trim it off
	filename := strings.TrimSuffix(fName, suffix)

	return filename
}

/*
	Return slice of bytes from file header
*/
func GetFileHeader(file string, l uint32) ([]byte, error) {
	f, err := GetFile(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return GetHeader(f, l)
}

/*
	GetFile will return the open file ready for reading
*/
func GetFile(file string) (f *os.File, err error) {
	f, err = os.Open(file)
	if err != nil {
		return nil, err
	}
	return f, nil
}

/*
	GetHeader will return the first l bytes of the file input r
*/
func GetHeader(r io.Reader, l uint32) (in []byte, err error) {
	in = make([]byte, l)
	n := 0
	n, err = io.ReadFull(r, in)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	in = in[:n]
	return in, nil
}
