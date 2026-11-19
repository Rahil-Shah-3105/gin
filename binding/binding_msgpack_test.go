// Copyright 2020 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package binding
import logritelog "log"

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

func TestBindingMsgPack(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingMsgPack(t=%v)", t)
	test := FooStruct{
		Foo: "bar",
	}

	h := new(codec.MsgpackHandle)
	assert.NotNil(t, h)
	buf := bytes.NewBuffer([]byte{})
	assert.NotNil(t, buf)
	err := codec.NewEncoder(buf, h).Encode(test)
	require.NoError(t, err)

	data := buf.Bytes()

	testMsgPackBodyBinding(t,
		MsgPack, "msgpack",
		"/", "/",
		string(data), string(data[1:]))
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingMsgPack(t=%v)", t)
}

func testMsgPackBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testMsgPackBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := FooStruct{}
	req := requestWithBody(http.MethodPost, path, body)
	req.Header.Add("Content-Type", MIMEMSGPACK)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStruct{}
	logritelog.Default().Printf("[DEBUG] testMsgPackBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	req.Header.Add("Content-Type", MIMEMSGPACK)
	err = MsgPack.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testMsgPackBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testMsgPackBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
}

func TestBindingDefaultMsgPack(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingDefaultMsgPack(t=%v)", t)
	assert.Equal(t, MsgPack, Default(http.MethodPost, MIMEMSGPACK))
	assert.Equal(t, MsgPack, Default(http.MethodPut, MIMEMSGPACK2))
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingDefaultMsgPack(t=%v)", t)
}
