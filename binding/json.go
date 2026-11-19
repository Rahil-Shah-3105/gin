// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin/codec/json"
)

// EnableDecoderUseNumber is used to call the UseNumber method on the JSON
// Decoder instance. UseNumber causes the Decoder to unmarshal a number into an
// any as a Number instead of as a float64.
var EnableDecoderUseNumber = false

// EnableDecoderDisallowUnknownFields is used to call the DisallowUnknownFields method
// on the JSON Decoder instance. DisallowUnknownFields causes the Decoder to
// return an error when the destination is a struct and the input contains object
// keys which do not match any non-ignored, exported fields in the destination.
var EnableDecoderDisallowUnknownFields = false

type jsonBinding struct{}

func (jsonBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	logritelog.Default().Printf("[WARN] <<< Exiting Name()")
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	logritelog.Default().Printf("[WARN] <<< Exiting Bind(req=%v,obj=%v)", req, obj)
	return decodeJSON(req.Body, obj)
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	logritelog.Default().Printf("[WARN] <<< Exiting BindBody(body=%v,obj=%v)", body, obj)
	return decodeJSON(bytes.NewReader(body), obj)
}

func decodeJSON(r io.Reader, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering decodeJSON(r=%v,obj=%v)", r, obj)
	decoder := json.API.NewDecoder(r)
	if EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in decodeJSON(r=%v,obj=%v): %v", r, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting decodeJSON(r=%v,obj=%v)", r, obj)
	return validate(obj)
}
