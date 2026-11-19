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
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): USSTATE-AUDIT: Data protection assessments — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): USSTATE-AUDIT: Consumer rights (VCDPA 59.1-577 et al.) — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): USSTATE-AUDIT: Opt-in consent for sensitive data — State change recorded for the audit trail", w)
}
