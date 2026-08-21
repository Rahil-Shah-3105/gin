// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http"

	"github.com/pelletier/go-toml/v2"
)

// TOML contains the given interface object.
type TOML struct {
	Data any
}

var tomlContentType = []string{"application/toml; charset=utf-8"}

// Render (TOML) marshals the given interface object and writes data with custom ContentType.
func (r TOML) Render(w http.ResponseWriter) error {
	logritelog.Default().Printf(">>> Entering Render(w=%v)", w)
	r.WriteContentType(w)

	bytes, err := toml.Marshal(r.Data)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Render(w=%v): %v", w, err)
		return err
	}

	_, err = w.Write(bytes)
	logritelog.Default().Printf("<<< Exiting Render(w=%v)", w)
	return err
}

// WriteContentType (TOML) writes TOML ContentType for response.
func (r TOML) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, tomlContentType)
	logritelog.Default().Printf("<<< Exiting WriteContentType(w=%v)", w)
}
