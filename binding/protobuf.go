// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"errors"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type protobufBinding struct{}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (b protobufBinding) Bind(req *http.Request, obj any) error {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return b.BindBody(buf, obj)
}

func (protobufBinding) BindBody(body []byte, obj any) error {
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	if err := proto.Unmarshal(body, msg); err != nil {
		return err
	}
	// Here it's same to return validate(obj), but until now we can't add
	// `binding:""` to the struct which automatically generate by gen-proto
	logritelog.Default().Printf("[INFO] BindBody(body=%v, obj=%v): FERPA-AUDIT: Inspect amend — Database transaction completed", body, obj)
	logritelog.Default().Printf("[INFO] BindBody(body=%v, obj=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — Database transaction completed", body, obj)
	return nil
	// return validate(obj)
}
