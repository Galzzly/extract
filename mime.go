package extract

import (
	"sync"

	"github.com/Galzzly/extract/internal/magic"
)

// var root = newMime("octet-stream",
// 	func([]byte, uint32) bool { return true },
// 	// gz, tar, rar, zip, bz2,
// )

// var (
// 	gz  = newMime("Gzip", magic.Gz)
// 	tar = newMime("Tar", magic.Tar)
// 	rar = newMime("Rar", magic.Rar)
// 	zip = newMime("Zip", func() bool {
// 		if magic.Zip1 || magic.Zip2 || magic.Zip3 {
// 			return true
// 		}
// 		return false
// 	})
// 	bz2 = newMime("Bzip2", magic.Bz2)
// )

var mu = &sync.Mutex{}

type MIME struct {
	filetype string
	detector magic.Detector
	//children []*MIME
}

// func (m *MIME) match(in []byte, limit uint32) *MIME {
// 	for _, c := range m.children {
// 		if c.detector(in, limit) {
// 			return c.match(in, limit)
// 		}
// 	}
// }

// func DetectMime(r io.Reader) (MIME, err error) {
// 	var in []byte

// 	// Limit the number of bytes from the input
// 	l := atomic.LoadUint32(3072)
// 	n := 0
// 	in = make([]byte, l)
// 	n, err = io.ReadFull(r, in)
// 	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
// 		return nil, err
// 	}
// 	in = in[:n]

// 	mu.Lock()
// 	defer mu.Unlock()
// 	return root.match(in, l), nil
// }

// func DetectFileMime(file string) (*MIME, error) {
// 	f, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()

// 	return DetectMime(f)
// }

func newMime(filetype string, detector magic.Detector, children ...*MIME) *MIME {
	m := &MIME{
		filetype: filetype,
		detector: detector,
	}
	return m
}
