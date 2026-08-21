// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	logritelog.Default().Printf(">>> Entering Name()")
	return "xml"
}

func (xmlBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf(">>> Entering Bind(req=%v,obj=%v)", req, obj)
	return decodeXML(req.Body, obj)
}

func (xmlBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf(">>> Entering BindBody(body=%v,obj=%v)", body, obj)
	return decodeXML(bytes.NewReader(body), obj)
}

func decodeXML(r io.Reader, obj any) error {
	logritelog.Default().Printf(">>> Entering decodeXML(r=%v,obj=%v)", r, obj)
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
	logritelog.Default().Printf("<<< Exiting decodeXML(r=%v,obj=%v)", r, obj)
		return err
	}
	return validate(obj)
}
