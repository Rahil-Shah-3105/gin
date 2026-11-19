// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http"
	"strconv"
)

// Data contains ContentType and bytes data.
type Data struct {
	ContentType string
	Data        []byte
}

// Render (Data) writes data with custom ContentType.
func (r Data) Render(w http.ResponseWriter) (err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Render(w=%v)", w)
	r.WriteContentType(w)
	if len(r.Data) > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(len(r.Data)))
	}
	_, err = w.Write(r.Data)
	logritelog.Default().Printf("[WARN] <<< Exiting Render(w=%v)", w)
	return
}

// WriteContentType (Data) writes custom ContentType.
func (r Data) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf("[ERROR] >>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, []string{r.ContentType})
	logritelog.Default().Printf("[WARN] <<< Exiting WriteContentType(w=%v)", w)
}
