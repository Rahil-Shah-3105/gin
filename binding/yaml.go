// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"bytes"
	"io"
	"net/http"

	"github.com/goccy/go-yaml"
)

type yamlBinding struct{}

func (yamlBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	return "yaml"
}

func (yamlBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	return decodeYAML(req.Body, obj)
}

func (yamlBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	return decodeYAML(bytes.NewReader(body), obj)
}

func decodeYAML(r io.Reader, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering decodeYAML(r=%v,obj=%v)", r, obj)
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in decodeYAML(r=%v,obj=%v): %v", r, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting decodeYAML(r=%v,obj=%v)", r, obj)
	return validate(obj)
}
