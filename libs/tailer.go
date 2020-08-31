package csd

// This file contains utilities for tailing input stream


import (
	"io"
	"os"
	"time"
)

var FileFollowPollInterval = 5 * time.Second

type followReader struct {
	f      *os.File
	follow bool
}

func NewFollowReader(fname string, follow bool) (*followReader, error) {
	var err error
	f := &followReader{}
	f.f, err = os.Open(fname)
	if err != nil {
		return nil, err
	}
	f.follow = follow
	return f, nil
}

func (f *followReader) Read(p []byte) (int, error) {
	for {
		n, err := f.f.Read(p)
		if err == nil {
			return n, err
		}
		if f.follow && err == io.EOF {
			time.Sleep(FileFollowPollInterval)
		} else {
			return n, err
		}
	}
}

func (f *followReader) Close() {
	f.f.Close()
}
