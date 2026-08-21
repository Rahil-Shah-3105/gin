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
	logritelog.Default().Printf(">>> Entering Render(w=%v)", w)
	r.WriteContentType(w)

	bytes, err := yaml.Marshal(r.Data)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Render(w=%v): %v", w, err)
		return err
	}

	_, err = w.Write(bytes)
	logritelog.Default().Printf("<<< Exiting Render(w=%v)", w)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r YAML) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, yamlContentType)
	logritelog.Default().Printf("<<< Exiting WriteContentType(w=%v)", w)
}
