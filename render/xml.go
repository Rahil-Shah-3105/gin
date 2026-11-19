// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"encoding/xml"
	"net/http"
)

// XML contains the given interface object.
type XML struct {
	Data any
}

var xmlContentType = []string{"application/xml; charset=utf-8"}

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (r XML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return xml.NewEncoder(w).Encode(r.Data)
}

// WriteContentType (XML) writes XML ContentType for response.
func (r XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, xmlContentType)
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 20]: Governance — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT [Art. 21(2)(c)]: Business continuity — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): NIS2-AUDIT: Security events — State change recorded for the audit trail", w)
}
