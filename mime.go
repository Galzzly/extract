package extract

import (
	"sync"

	"github.com/Galzzly/extract/internal/magic"
)

var mu = &sync.Mutex{}

type MIME struct {
	filetype string
	detector magic.Detector
	//children []*MIME
}

func newMime(filetype string, detector magic.Detector, children ...*MIME) *MIME {
	m := &MIME{
		filetype: filetype,
		detector: detector,
	}
	return m
}
