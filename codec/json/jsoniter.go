// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build jsoniter

package json
import logritelog "log"

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

// Package indicates what library is being used for JSON encoding.
const Package = "github.com/json-iterator/go"

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	API = jsoniterApi{}
	logritelog.Default().Printf("[DEBUG] init(): API → %v", API)
logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type jsoniterApi struct{}

func (j jsoniterApi) Marshal(v any) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Marshal(v=%v)", v)
	return json.Marshal(v)
}

func (j jsoniterApi) Unmarshal(data []byte, v any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Unmarshal(data=%v,v=%v)", data, v)
	return json.Unmarshal(data, v)
}

func (j jsoniterApi) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalIndent(v=%v,prefix=%v,indent=%v)", v, prefix, indent)
	return json.MarshalIndent(v, prefix, indent)
}

func (j jsoniterApi) NewEncoder(writer io.Writer) Encoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewEncoder(writer=%v)", writer)
	return json.NewEncoder(writer)
}

func (j jsoniterApi) NewDecoder(reader io.Reader) Decoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewDecoder(reader=%v)", reader)
	return json.NewDecoder(reader)
}
