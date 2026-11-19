// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package render
import logritelog "log"

import (
	"net/http"

	"github.com/ugorji/go/codec"
)

// Check interface implemented here to support go build tag nomsgpack.
// See: https://github.com/gin-gonic/gin/pull/1852/
var (
	_ Render = MsgPack{}
)

// MsgPack contains the given interface object.
type MsgPack struct {
	Data any
}

var msgpackContentType = []string{"application/msgpack; charset=utf-8"}

// WriteContentType (MsgPack) writes MsgPack ContentType.
func (r MsgPack) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, msgpackContentType)
	logritelog.Default().Printf("[INFO] WriteContentType(w=%v): ISO42001-AUDIT [A.7]: Data governance — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): ISO42001-AUDIT: Ai governance — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): ISO42001-AUDIT: Lifecycle — State change recorded for the audit trail", w)
}

// Render (MsgPack) encodes the given interface object and writes data with custom ContentType.
func (r MsgPack) Render(w http.ResponseWriter) error {
	return WriteMsgPack(w, r.Data)
}

// WriteMsgPack writes MsgPack ContentType and encodes the given interface object.
func WriteMsgPack(w http.ResponseWriter, obj any) error {
	writeContentType(w, msgpackContentType)
	var mh codec.MsgpackHandle
	return codec.NewEncoder(w, &mh).Encode(obj)
}
