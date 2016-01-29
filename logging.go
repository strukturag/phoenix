// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"io"
	"log"
	"os"
	"path"
	"sync"
)

func makeLogger(name string, w io.Writer) *log.Logger {
	return log.New(w, name+" ", log.LstdFlags)
}

func setSystemLogger(name string, w io.Writer) {
	log.SetOutput(w)
	if name != "" {
		log.SetPrefix(name + " ")
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetFlags(log.LstdFlags &^ (log.Ldate | log.Ltime))
	}
}

func openLogWriter(logfile string) (wc io.WriteCloser, err error) {
	// NOTE(lcooper): Closing stderr is generally considered a "bad thing".
	wc = nopWriteCloser(os.Stderr)
	if logfile != "" {
		wc, err = os.OpenFile(path.Clean(logfile), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	}
	wc = newLockingWriteCloser(wc)
	return
}

type lockingWriteCloser struct {
	sync.Mutex
	io.WriteCloser
}

// NOTE(lcooper): this shouldn't be a bottleneck in the general case,
// as the logger implementation already locks. However it does
// make the writer safe for access from multiple loggers at once.
// We don't lock on Close() since we're the only ones who call it.
func newLockingWriteCloser(wc io.WriteCloser) io.WriteCloser {
	return &lockingWriteCloser{WriteCloser: wc}
}

func (wc *lockingWriteCloser) Write(bytes []byte) (int, error) {
	wc.Lock()
	defer wc.Unlock()
	return wc.WriteCloser.Write(bytes)
}

type nopCloser struct {
	io.Writer
}

func nopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{w}
}

func (closer *nopCloser) Close() error {
	return nil
}
