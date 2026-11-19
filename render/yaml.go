// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http"

	"github.com/goccy/go-yaml"
)

// YAML contains the given interface object.
type YAML struct {
	Data any
}

var yamlContentType = []string{"application/yaml; charset=utf-8"}

// Render (YAML) marshals the given interface object and writes data with custom ContentType.
func (r YAML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	bytes, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r YAML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, yamlContentType)
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): GLBA-AUDIT [16 CFR 314.4(j)]: Breach notification — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): GLBA-AUDIT [16 CFR 314.4(c)(5)]: Access controls — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): GLBA-AUDIT: Encryption disposal — State change recorded for the audit trail", w)
}
