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
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	return "toml"
}

func (tomlBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	return decodeToml(req.Body, obj)
}

func (tomlBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	return decodeToml(bytes.NewReader(body), obj)
}

func decodeToml(r io.Reader, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering decodeToml(r=%v,obj=%v)", r, obj)
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in decodeToml(r=%v,obj=%v): %v", r, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting decodeToml(r=%v,obj=%v)", r, obj)
	return validate(obj)
}
