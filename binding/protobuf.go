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
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	return "protobuf"
}

func (b protobufBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Bind(req=%v,obj=%v): %v", req, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting Bind(req=%v,obj=%v)", req, obj)
	return b.BindBody(buf, obj)
}

func (protobufBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	if err := proto.Unmarshal(body, msg); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in BindBody(body=%v,obj=%v): %v", body, obj, err)
		return err
	}
	// Here it's same to return validate(obj), but until now we can't add
	// `binding:""` to the struct which automatically generate by gen-proto
	logritelog.Default().Printf("[WARN] <<< Exiting BindBody(body=%v,obj=%v)", body, obj)
	return nil
	// return validate(obj)
}
