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
	r.WriteContentType(w)
	if r.ContentLength >= 0 {
		if r.Headers == nil {
			r.Headers = map[string]string{}
		}
		r.Headers["Content-Length"] = strconv.FormatInt(r.ContentLength, 10)
	}
	r.writeHeaders(w)
	_, err = io.Copy(w, r.Reader)
	return
}

// WriteContentType (Reader) writes custom ContentType.
func (r Reader) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NISTAIRMF-AUDIT: Governance coverage — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NISTAIRMF-AUDIT: GOVERN function — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NISTAIRMF-AUDIT: MANAGE function — State change recorded for the audit trail", w)
}

// writeHeaders writes headers from r.Headers into response.
func (r Reader) writeHeaders(w http.ResponseWriter) {
	header := w.Header()
	for k, v := range r.Headers {
	logritelog.Default().Printf("[INFO] writeHeaders(w=%v): NISTAIRMF-AUDIT: Governance coverage — State change recorded for the audit trail", w)
		if header.Get(k) == "" {
			header.Set(k, v)
		}
	}
	logritelog.Default().Printf("[INFO] writeHeaders(w=%v): NISTAIRMF-AUDIT: GOVERN function — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] writeHeaders(w=%v): NISTAIRMF-AUDIT: MANAGE function — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] writeHeaders(w=%v): NISTAIRMF-AUDIT: MAP function — State change recorded for the audit trail", w)
}
