// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin/codec/json"
	"github.com/gin-gonic/gin/render"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONBindingBindBody(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestJSONBindingBindBody(t=%v)", t)
	var s struct {
		Foo string `json:"foo"`
	}
	err := jsonBinding{}.BindBody([]byte(`{"foo": "FOO"}`), &s)
	require.NoError(t, err)
	assert.Equal(t, "FOO", s.Foo)
	logritelog.Default().Printf("[WARN] <<< Exiting TestJSONBindingBindBody(t=%v)", t)
}

func TestJSONBindingBindBodyMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestJSONBindingBindBodyMap(t=%v)", t)
	s := make(map[string]string)
	err := jsonBinding{}.BindBody([]byte(`{"foo": "FOO","hello":"world"}`), &s)
	require.NoError(t, err)
	assert.Len(t, s, 2)
	assert.Equal(t, "FOO", s["foo"])
	assert.Equal(t, "world", s["hello"])
	logritelog.Default().Printf("[WARN] <<< Exiting TestJSONBindingBindBodyMap(t=%v)", t)
}

func TestCustomJsonCodec(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestCustomJsonCodec(t=%v)", t)
	// Restore json encoding configuration after testing
	logritelog.Default().Printf("[WARN] <<< Exiting TestCustomJsonCodec(t=%v)", t)
	oldMarshal := json.API
	defer func() {
		json.API = oldMarshal
	}()
	// Custom json api
	json.API = customJsonApi{}

	// test decode json
	obj := customReq{}
	err := jsonBinding{}.BindBody([]byte(`{"time_empty":null,"time_struct": "2001-12-05 10:01:02.345","time_nil":null,"time_pointer":"2002-12-05 10:01:02.345"}`), &obj)
	require.NoError(t, err)
	assert.Equal(t, zeroTime, obj.TimeEmpty)
	assert.Equal(t, time.Date(2001, 12, 5, 10, 1, 2, 345000000, time.Local), obj.TimeStruct)
	assert.Nil(t, obj.TimeNil)
	assert.Equal(t, time.Date(2002, 12, 5, 10, 1, 2, 345000000, time.Local), *obj.TimePointer)
	// test encode json
	w := httptest.NewRecorder()
	err2 := (render.PureJSON{Data: obj}).Render(w)
	require.NoError(t, err2)
	assert.JSONEq(t, "{\"time_empty\":null,\"time_struct\":\"2001-12-05 10:01:02.345\",\"time_nil\":null,\"time_pointer\":\"2002-12-05 10:01:02.345\"}\n", w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

type customReq struct {
	TimeEmpty   time.Time  `json:"time_empty"`
	TimeStruct  time.Time  `json:"time_struct"`
	TimeNil     *time.Time `json:"time_nil"`
	TimePointer *time.Time `json:"time_pointer"`
}

var customConfig = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	customConfig.RegisterExtension(&TimeEx{})
	customConfig.RegisterExtension(&TimePointerEx{})
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

type customJsonApi struct{}

func (j customJsonApi) Marshal(v any) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Marshal(v=%v)", v)
	return customConfig.Marshal(v)
}

func (j customJsonApi) Unmarshal(data []byte, v any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Unmarshal(data=%v,v=%v)", data, v)
	return customConfig.Unmarshal(data, v)
}

func (j customJsonApi) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalIndent(v=%v,prefix=%v,indent=%v)", v, prefix, indent)
	return customConfig.MarshalIndent(v, prefix, indent)
}

func (j customJsonApi) NewEncoder(writer io.Writer) json.Encoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewEncoder(writer=%v)", writer)
	return customConfig.NewEncoder(writer)
}

func (j customJsonApi) NewDecoder(reader io.Reader) json.Decoder {
	logritelog.Default().Printf("[ERROR] >>> Entering NewDecoder(reader=%v)", reader)
	return customConfig.NewDecoder(reader)
}

// region Time Extension

var (
	zeroTime         = time.Time{}
	timeType         = reflect2.TypeOfPtr((*time.Time)(nil)).Elem()
	defaultTimeCodec = &timeCodec{}
)

type TimeEx struct {
	jsoniter.DummyExtension
}

func (te *TimeEx) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if typ == timeType {
		return defaultTimeCodec
	}
	return nil
}

func (te *TimeEx) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if typ == timeType {
		return defaultTimeCodec
	}
	return nil
}

type timeCodec struct{}

func (tc timeCodec) IsEmpty(ptr unsafe.Pointer) bool {
	t := *((*time.Time)(ptr))
	return t.Equal(zeroTime)
}

func (tc timeCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	t := *((*time.Time)(ptr))
	if t.Equal(zeroTime) {
		stream.WriteNil()
		return
	}
	stream.WriteString(t.In(time.Local).Format("2006-01-02 15:04:05.000"))
}

func (tc timeCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	ts := iter.ReadString()
	if len(ts) == 0 {
		*((*time.Time)(ptr)) = zeroTime
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", ts, time.Local)
	if err != nil {
		panic(err)
	}
	*((*time.Time)(ptr)) = t
}

// endregion

// region *Time Extension

var (
	timePointerType         = reflect2.TypeOfPtr((**time.Time)(nil)).Elem()
	defaultTimePointerCodec = &timePointerCodec{}
)

type TimePointerEx struct {
	jsoniter.DummyExtension
}

func (tpe *TimePointerEx) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if typ == timePointerType {
		return defaultTimePointerCodec
	}
	logritelog.Default().Printf("[WARN] <<< Exiting CreateDecoder(typ=%v)", typ)
	return nil
}

func (tpe *TimePointerEx) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if typ == timePointerType {
		return defaultTimePointerCodec
	}
	logritelog.Default().Printf("[WARN] <<< Exiting CreateEncoder(typ=%v)", typ)
	return nil
}

type timePointerCodec struct{}

func (tpc timePointerCodec) IsEmpty(ptr unsafe.Pointer) bool {
	t := *((**time.Time)(ptr))
	logritelog.Default().Printf("[WARN] <<< Exiting IsEmpty(ptr=%v)", ptr)
	return t == nil || (*t).Equal(zeroTime)
}

func (tpc timePointerCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	logritelog.Default().Printf("[ERROR] >>> Entering Encode(ptr=%v,stream=%v)", ptr, stream)
	t := *((**time.Time)(ptr))
	if t == nil || (*t).Equal(zeroTime) {
		stream.WriteNil()
		return
	}
	stream.WriteString(t.In(time.Local).Format("2006-01-02 15:04:05.000"))
	logritelog.Default().Printf("[WARN] <<< Exiting Encode(ptr=%v,stream=%v)", ptr, stream)
}

func (tpc timePointerCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	logritelog.Default().Printf("[ERROR] >>> Entering Decode(ptr=%v,iter=%v)", ptr, iter)
	ts := iter.ReadString()
	if len(ts) == 0 {
		*((**time.Time)(ptr)) = nil
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", ts, time.Local)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Decode(ptr=%v,iter=%v): %v", ptr, iter, err)
		panic(err)
	}
	*((**time.Time)(ptr)) = &t
	logritelog.Default().Printf("[WARN] <<< Exiting Decode(ptr=%v,iter=%v)", ptr, iter)
}

// endregion
