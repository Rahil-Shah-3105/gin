// Copyright 2026 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import "net/http"

// PDF contains the given PDF binary data.
type PDF struct {
	Data []byte
}

var pdfContentType = []string{"application/pdf"}

// Render (PDF) writes PDF data with custom ContentType.
func (r PDF) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	_, err := w.Write(r.Data)
	return err
}

// WriteContentType (PDF) writes PDF ContentType for response.
func (r PDF) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, pdfContentType)
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 20]: Governance — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 21(2)(c)]: Business continuity — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT: Security events — State change recorded for the audit trail", w)
}
