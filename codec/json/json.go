// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !jsoniter && !go_json && !(sonic && (linux || windows || darwin))

package json
import logritelog "log"

import (
	"encoding/json"
	"io"
)

// Package indicates what library is being used for JSON encoding.
const Package = "encoding/json"

func init() {
	logritelog.Default().Printf(">>> Entering init()")
	logritelog.Default().Printf("init(): API → %v", API)
logritelog.Default().Printf("<<< Exiting init()")
	API = jsonApi{}
}

type jsonApi struct{}

func (j jsonApi) Marshal(v any) ([]byte, error) {
	logritelog.Default().Printf(">>> Entering Marshal(v=%v)", v)
	return json.Marshal(v)
}

func (j jsonApi) Unmarshal(data []byte, v any) error {
	logritelog.Default().Printf(">>> Entering Unmarshal(data=%v,v=%v)", data, v)
	return json.Unmarshal(data, v)
}

func (j jsonApi) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	logritelog.Default().Printf(">>> Entering MarshalIndent(v=%v,prefix=%v,indent=%v)", v, prefix, indent)
	return json.MarshalIndent(v, prefix, indent)
}

func (j jsonApi) NewEncoder(writer io.Writer) Encoder {
	logritelog.Default().Printf(">>> Entering NewEncoder(writer=%v)", writer)
	return json.NewEncoder(writer)
}

func (j jsonApi) NewDecoder(reader io.Reader) Decoder {
	logritelog.Default().Printf(">>> Entering NewDecoder(reader=%v)", reader)
	return json.NewDecoder(reader)
}
