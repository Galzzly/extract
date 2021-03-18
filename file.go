package extract

import (
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
