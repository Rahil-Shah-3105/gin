// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package binding
import logritelog "log"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

func TestMsgpackBindingBindBody(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMsgpackBindingBindBody(t=%v)", t)
	type teststruct struct {
		Foo string `msgpack:"foo"`
	}
	var s teststruct
	err := msgpackBinding{}.BindBody(msgpackBody(t, teststruct{"FOO"}), &s)
	require.NoError(t, err)
	assert.Equal(t, "FOO", s.Foo)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMsgpackBindingBindBody(t=%v)", t)
}

func msgpackBody(t *testing.T, obj any) []byte {
	logritelog.Default().Printf("[ERROR] >>> Entering msgpackBody(t=%v,obj=%v)", t, obj)
	var bs bytes.Buffer
	h := &codec.MsgpackHandle{}
	err := codec.NewEncoder(&bs, h).Encode(obj)
	require.NoError(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting msgpackBody(t=%v,obj=%v)", t, obj)
	return bs.Bytes()
}
