// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package binding
import logritelog "log"

import (
	"bytes"
	"io"
	"net/http"

	"github.com/ugorji/go/codec"
)

type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	logritelog.Default().Printf("[WARN] <<< Exiting Name()")
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	logritelog.Default().Printf("[WARN] <<< Exiting Bind(req=%v,obj=%v)", req, obj)
	return decodeMsgPack(req.Body, obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	logritelog.Default().Printf("[WARN] <<< Exiting BindBody(body=%v,obj=%v)", body, obj)
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering decodeMsgPack(r=%v,obj=%v)", r, obj)
	cdc := new(codec.MsgpackHandle)
	if err := codec.NewDecoder(r, cdc).Decode(&obj); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in decodeMsgPack(r=%v,obj=%v): %v", r, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting decodeMsgPack(r=%v,obj=%v)", r, obj)
	return validate(obj)
}
