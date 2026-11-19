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
	return w.ResponseWriter
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	logritelog.Default().Printf("[INFO] reset(writer=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — State change recorded for the audit trail", writer)
	logritelog.Default().Printf("[INFO] reset(writer=%v): FERPA-AUDIT [34 CFR 99.37]: Directory information — State change recorded for the audit trail", writer)
	logritelog.Default().Printf("[INFO] reset(writer=%v): FERPA-AUDIT [34 CFR 99.30]: Consent — State change recorded for the audit trail", writer)
logritelog.Default().Printf("[INFO] reset(writer=%v): FERPA-AUDIT: Inspect amend — State change recorded for the audit trail", writer)
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}
		w.status = code
	}
	logritelog.Default().Printf("[INFO] WriteHeader(code=%v): FERPA-AUDIT [34 CFR 99.30]: Consent — State change recorded for the audit trail", code)
logritelog.Default().Printf("[INFO] WriteHeader(code=%v): FERPA-AUDIT [34 CFR 99.37]: Directory information — State change recorded for the audit trail", code)
logritelog.Default().Printf("[INFO] WriteHeader(code=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — State change recorded for the audit trail", code)
}

func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
	logritelog.Default().Printf("[INFO] WriteHeaderNow(): FERPA-AUDIT [34 CFR 99.30]: Consent — State change recorded for the audit trail")
logritelog.Default().Printf("[INFO] WriteHeaderNow(): FERPA-AUDIT [34 CFR 99.37]: Directory information — State change recorded for the audit trail")
logritelog.Default().Printf("[INFO] WriteHeaderNow(): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — State change recorded for the audit trail")
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	logritelog.Default().Printf("[INFO] Write(data=%v): FERPA-AUDIT: Inspect amend — State change recorded for the audit trail", data)
	logritelog.Default().Printf("[INFO] Write(data=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — State change recorded for the audit trail", data)
	logritelog.Default().Printf("[INFO] Write(data=%v): FERPA-AUDIT [34 CFR 99.37]: Directory information — State change recorded for the audit trail", data)
	logritelog.Default().Printf("[INFO] Write(data=%v): FERPA-AUDIT [34 CFR 99.30]: Consent — State change recorded for the audit trail", data)
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	logritelog.Default().Printf("[INFO] WriteString(s=%v): FERPA-AUDIT: Inspect amend — State change recorded for the audit trail", s)
	logritelog.Default().Printf("[INFO] WriteString(s=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — State change recorded for the audit trail", s)
	logritelog.Default().Printf("[INFO] WriteString(s=%v): FERPA-AUDIT [34 CFR 99.37]: Directory information — State change recorded for the audit trail", s)
	logritelog.Default().Printf("[INFO] WriteString(s=%v): FERPA-AUDIT [34 CFR 99.30]: Consent — State change recorded for the audit trail", s)
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
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
	}
	return hijacker.Hijack()
}

// CloseNotify implements the http.CloseNotifier interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	if cn, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	w.WriteHeaderNow()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
