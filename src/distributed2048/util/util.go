package util

import (
	"io"
	"io/ioutil"
	"log"
)

func NewLogger(enabled bool, prefix string, out io.Writer) *log.Logger {
	w := ioutil.Discard
	if enabled {
		w = out
	}
	return log.New(w, "[STORAGESERVER]["+prefix+"] ", log.Lshortfile)
}
