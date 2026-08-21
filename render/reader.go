// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"io"
	"net/http"
	"strconv"
)

// Reader contains the IO reader and its length, and custom ContentType and other headers.
type Reader struct {
	ContentType   string
	ContentLength int64
	Reader        io.Reader
	Headers       map[string]string
}

// Render (Reader) writes data with custom ContentType and headers.
func (r Reader) Render(w http.ResponseWriter) (err error) {
	logritelog.Default().Printf(">>> Entering Render(w=%v)", w)
	r.WriteContentType(w)
	if r.ContentLength >= 0 {
		if r.Headers == nil {
			r.Headers = map[string]string{}
			logritelog.Default().Printf("Render(w=%v): r.Headers → %v", w, r.Headers)
		}
		r.Headers["Content-Length"] = strconv.FormatInt(r.ContentLength, 10)
	}
	r.writeHeaders(w)
	_, err = io.Copy(w, r.Reader)
	logritelog.Default().Printf("<<< Exiting Render(w=%v)", w)
	return
}

// WriteContentType (Reader) writes custom ContentType.
func (r Reader) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, []string{r.ContentType})
	logritelog.Default().Printf("<<< Exiting WriteContentType(w=%v)", w)
}

// writeHeaders writes headers from r.Headers into response.
func (r Reader) writeHeaders(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering writeHeaders(w=%v)", w)
	header := w.Header()
	for k, v := range r.Headers {
		if header.Get(k) == "" {
			header.Set(k, v)
		}
	}
	logritelog.Default().Printf("<<< Exiting writeHeaders(w=%v)", w)
}
