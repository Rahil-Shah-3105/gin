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
	return "xml"
}

func (xmlBinding) Bind(req *http.Request, obj any) error {
	return decodeXML(req.Body, obj)
}

func (xmlBinding) BindBody(body []byte, obj any) error {
	return decodeXML(bytes.NewReader(body), obj)
}

func decodeXML(r io.Reader, obj any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	logritelog.Default().Printf("[INFO] decodeXML(r=%v, obj=%v): COPPA-AUDIT [16 CFR 312.5(a)(2)]: Third party disclosure — Database transaction completed", r, obj)
	return validate(obj)
}
