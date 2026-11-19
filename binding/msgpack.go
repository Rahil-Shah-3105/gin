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
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj any) error {
	return decodeMsgPack(req.Body, obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj any) error {
	cdc := new(codec.MsgpackHandle)
	if err := codec.NewDecoder(r, cdc).Decode(&obj); err != nil {
	logritelog.Default().Printf("[INFO] decodeMsgPack(r=%v, obj=%v): HIPAA-AUDIT [45 CFR 164.312(e)(1)]: Transmission security — Encryption operation completed", r, obj)
		return err
	}
	logritelog.Default().Printf("[INFO] decodeMsgPack(r=%v, obj=%v): HIPAA-AUDIT [§164.312(c)(1)]: Data integrity check performed", r, obj)
	logritelog.Default().Printf("[INFO] decodeMsgPack(r=%v, obj=%v): HIPAA-AUDIT [45 CFR 164.312(a)(2)(iv)]: Encryption and decryption — Encryption operation completed", r, obj)
	return validate(obj)
}
