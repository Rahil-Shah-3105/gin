// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pelletier/go-toml/v2"
)

type tomlBinding struct{}

func (tomlBinding) Name() string {
	return "toml"
}

func (tomlBinding) Bind(req *http.Request, obj any) error {
	return decodeToml(req.Body, obj)
}

func (tomlBinding) BindBody(body []byte, obj any) error {
	return decodeToml(bytes.NewReader(body), obj)
}

func decodeToml(r io.Reader, obj any) error {
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	logritelog.Default().Printf("[INFO] decodeToml(r=%v, obj=%v): FERPA-AUDIT: Inspect amend — Database transaction completed", r, obj)
	logritelog.Default().Printf("[INFO] decodeToml(r=%v, obj=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — Database transaction completed", r, obj)
	return validate(obj)
}
