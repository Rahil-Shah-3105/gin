// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

var errHijackAlreadyWritten = errors.New("gin: response body already written")

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = (*responseWriter)(nil)

func (w *responseWriter) Unwrap() http.ResponseWriter {
	logritelog.Default().Printf(">>> Entering Unwrap()")
	return w.ResponseWriter
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering reset(writer=%v)", writer)
	w.ResponseWriter = writer
	w.size = noWritten
	logritelog.Default().Printf("reset(writer=%v): w.size → %v", writer, w.size)
	logritelog.Default().Printf("reset(writer=%v): w.status → %v", writer, w.status)
logritelog.Default().Printf("reset(writer=%v): w.ResponseWriter → %v", writer, w.ResponseWriter)
logritelog.Default().Printf("<<< Exiting reset(writer=%v)", writer)
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	logritelog.Default().Printf(">>> Entering WriteHeader(code=%v)", code)
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}
		w.status = code
		logritelog.Default().Printf("WriteHeader(code=%v): w.status → %v", code, w.status)
	}
	logritelog.Default().Printf("<<< Exiting WriteHeader(code=%v)", code)
}

func (w *responseWriter) WriteHeaderNow() {
	logritelog.Default().Printf(">>> Entering WriteHeaderNow()")
	if !w.Written() {
		w.size = 0
		logritelog.Default().Printf("WriteHeaderNow(): w.size → %v", w.size)
		w.ResponseWriter.WriteHeader(w.status)
	}
	logritelog.Default().Printf("<<< Exiting WriteHeaderNow()")
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	logritelog.Default().Printf(">>> Entering Write(data=%v)", data)
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	logritelog.Default().Printf("<<< Exiting Write(data=%v)", data)
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	logritelog.Default().Printf(">>> Entering WriteString(s=%v)", s)
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	logritelog.Default().Printf("<<< Exiting WriteString(s=%v)", s)
	return
}

func (w *responseWriter) Status() int {
	logritelog.Default().Printf(">>> Entering Status()")
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	logritelog.Default().Printf(">>> Entering Written()")
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	logritelog.Default().Printf(">>> Entering Hijack()")
	// Allow hijacking before any data is written (size == -1) or after headers are written (size == 0),
	// but not after body data is written (size > 0). For compatibility with websocket libraries (e.g., github.com/coder/websocket)
	if w.size > 0 {
		return nil, nil, errHijackAlreadyWritten
	}
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	if w.size < 0 {
		w.size = 0
		logritelog.Default().Printf("Hijack(): w.size → %v", w.size)
	}
	logritelog.Default().Printf("<<< Exiting Hijack()")
	return hijacker.Hijack()
}

// CloseNotify implements the http.CloseNotifier interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	logritelog.Default().Printf(">>> Entering CloseNotify()")
	if cn, ok := w.ResponseWriter.(http.CloseNotifier); ok {
	logritelog.Default().Printf("<<< Exiting CloseNotify()")
		return cn.CloseNotify()
	}
	return nil
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	logritelog.Default().Printf(">>> Entering Flush()")
	w.WriteHeaderNow()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	logritelog.Default().Printf("<<< Exiting Flush()")
}

func (w *responseWriter) Pusher() (pusher http.Pusher) {
	logritelog.Default().Printf(">>> Entering Pusher()")
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
	logritelog.Default().Printf("<<< Exiting Pusher()")
		return pusher
	}
	return nil
}
