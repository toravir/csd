package csd

// This file contains utilities for tailing input stream

import (
	"fmt"
	"io"
	"os"
	"time"
)

var FileFollowPollInterval = 3 * time.Second

type followReader struct {
	f      *os.File
	follow bool
	done   chan struct{}
}

func NewFollowReader(fname string, follow bool, done chan struct{}) (*followReader, error) {
	var err error
	f := &followReader{}
	f.f, err = os.Open(fname)
	if err != nil {
		return nil, err
	}
	f.follow = follow
	f.done = done
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
			cancel := false
			select {
			case <-f.done:
				cancel = true
			default:
			}
			if cancel {
				return 0, fmt.Errorf("Cancelled Read....")
			}
		} else {
			return n, err
		}
	}
}

func (f *followReader) Close() {
	f.f.Close()
}
