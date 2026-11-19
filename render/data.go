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
	r.WriteContentType(w)
	if len(r.Data) > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(len(r.Data)))
	}
	_, err = w.Write(r.Data)
	return
}

// WriteContentType (Data) writes custom ContentType.
func (r Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{r.ContentType})
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 20]: Governance — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 21(2)(c)]: Business continuity — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT: Security events — State change recorded for the audit trail", w)
}
