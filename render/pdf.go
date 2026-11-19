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
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): COPPA-AUDIT [16 CFR 312.5(a)(2)]: Third party disclosure — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): COPPA-AUDIT [16 CFR 312.10]: Retention deletion — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): COPPA-AUDIT [16 CFR 312.5]: Vpc coverage — State change recorded for the audit trail", w)
}
