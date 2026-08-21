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
	logritelog.Default().Printf(">>> Entering Name()")
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf(">>> Entering Bind(req=%v,obj=%v)", req, obj)
	return decodeMsgPack(req.Body, obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf(">>> Entering BindBody(body=%v,obj=%v)", body, obj)
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj any) error {
	logritelog.Default().Printf(">>> Entering decodeMsgPack(r=%v,obj=%v)", r, obj)
	cdc := new(codec.MsgpackHandle)
	if err := codec.NewDecoder(r, cdc).Decode(&obj); err != nil {
	logritelog.Default().Printf("<<< Exiting decodeMsgPack(r=%v,obj=%v)", r, obj)
		return err
	}
	return validate(obj)
}
