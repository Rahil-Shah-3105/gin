// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build sonic && (linux || windows || darwin)

package json
import logritelog "log"

import (
	"io"

	"github.com/bytedance/sonic"
)

// Package indicates what library is being used for JSON encoding.
const Package = "github.com/bytedance/sonic"

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	API = sonicApi{}
	logritelog.Default().Printf("[DEBUG] init(): API → %v", API)
logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

var json = sonic.ConfigStd

type sonicApi struct{}

func (j sonicApi) Marshal(v any) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Marshal(v=%v)", v)
	logritelog.Default().Printf("[WARN] <<< Exiting Marshal(v=%v)", v)
	return json.Marshal(v)
}

func (j sonicApi) Unmarshal(data []byte, v any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Unmarshal(data=%v,v=%v)", data, v)
	logritelog.Default().Printf("[WARN] <<< Exiting Unmarshal(data=%v,v=%v)", data, v)
	return json.Unmarshal(data, v)
}

func (j sonicApi) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalIndent(v=%v,prefix=%v,indent=%v)", v, prefix, indent)
	logritelog.Default().Printf("[WARN] <<< Exiting MarshalIndent(v=%v,prefix=%v,indent=%v)", v, prefix, indent)
	return json.MarshalIndent(v, prefix, indent)
}

func (j sonicApi) NewEncoder(writer io.Writer) Encoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewEncoder(writer=%v)", writer)
	logritelog.Default().Printf("[WARN] <<< Exiting NewEncoder(writer=%v)", writer)
	return json.NewEncoder(writer)
}

func (j sonicApi) NewDecoder(reader io.Reader) Decoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewDecoder(reader=%v)", reader)
	logritelog.Default().Printf("[WARN] <<< Exiting NewDecoder(reader=%v)", reader)
	return json.NewDecoder(reader)
}
