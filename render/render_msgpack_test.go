// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !nomsgpack

package render
import logritelog "log"

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

func TestRenderMsgPack(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderMsgPack(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(MsgPack{data}).WriteContentType(w)
	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))

	err := (MsgPack{data}).Render(w)

	require.NoError(t, err)

	var decoded map[string]any
	var mh codec.MsgpackHandle
	mh.RawToString = true
	logritelog.Default().Printf("[DEBUG] TestRenderMsgPack(t=%v): mh.RawToString → %v", t, mh.RawToString)
	err = codec.NewDecoderBytes(w.Body.Bytes(), &mh).Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, data, decoded)
	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[DEBUG] TestRenderMsgPack(t=%v): err → %v", t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderMsgPack(t=%v)", t)
}

func TestWriteMsgPack(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestWriteMsgPack(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
		"num": 42,
	}

	err := WriteMsgPack(w, data)
	require.NoError(t, err)

	assert.Equal(t, "application/msgpack; charset=utf-8", w.Header().Get("Content-Type"))

	var decoded map[string]any
	var mh codec.MsgpackHandle
	mh.RawToString = true
	logritelog.Default().Printf("[DEBUG] TestWriteMsgPack(t=%v): mh.RawToString → %v", t, mh.RawToString)
	err = codec.NewDecoderBytes(w.Body.Bytes(), &mh).Decode(&decoded)
	require.NoError(t, err)
	assert.Len(t, decoded, 2)
	assert.Equal(t, "bar", decoded["foo"])
	assert.EqualValues(t, 42, decoded["num"])
	logritelog.Default().Printf("[DEBUG] TestWriteMsgPack(t=%v): err → %v", t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestWriteMsgPack(t=%v)", t)
}

type failWriter struct {
	*httptest.ResponseRecorder
}

func (w *failWriter) Write(data []byte) (int, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Write(data=%v)", data)
	logritelog.Default().Printf("[WARN] <<< Exiting Write(data=%v)", data)
	return 0, errors.New("write error")
}

func TestRenderMsgPackError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderMsgPackError(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	err := (MsgPack{data}).Render(&failWriter{w})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderMsgPackError(t=%v)", t)
}
