package main

import (
	"fmt"
	"io"
	"time"
)

type timeoutReader struct {
	io.Reader
	dur time.Duration
}

func (t timeoutReader) Read(p []byte) (n int, err error) {
	out := make(chan struct{})

	go func() {
		n, err = t.Reader.Read(p)
		close(out)
	}()

	select {
	case <-time.After(t.dur):
		return n, fmt.Errorf("timed out waiting to read for %v", t.dur)
	case <-out:
		return
	}
}
