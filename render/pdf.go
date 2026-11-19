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
	logritelog.Default().Printf("[WARN] WriteContentType(w=%v): GDPR-AUDIT [Art. 6]: Lawfulness of processing — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): GDPR-AUDIT: Breach notification — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): GDPR-AUDIT [Art. 25]: Data protection by design and by default — State change recorded for the audit trail", w)
}
