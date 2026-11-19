// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import (
	logritelog "log"
	logritetime "time"
)

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin/testdata/protoexample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
)

type appkey struct {
	Appkey string `json:"appkey" form:"appkey"`
}

type QueryTest struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
	appkey
}

type FooStruct struct {
	Foo string `msgpack:"foo" json:"foo" form:"foo" xml:"foo" binding:"required,max=32"`
}

type FooBarStruct struct {
	FooStruct
	Bar string `msgpack:"bar" json:"bar" form:"bar" xml:"bar" binding:"required"`
}

type FooBarFileStruct struct {
	FooBarStruct
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type FooBarFileFailStruct struct {
	FooBarStruct
	File *multipart.FileHeader `invalid_name:"file" binding:"required"`
}

type FooDefaultBarStruct struct {
	FooStruct
	Bar string `msgpack:"bar" json:"bar" form:"bar,default=hello" xml:"bar" binding:"required"`
}

type FooStructUseNumber struct {
	Foo any `json:"foo" binding:"required"`
}

type FooStructDisallowUnknownFields struct {
	Foo any `json:"foo" binding:"required"`
}

type FooBarStructForTimeType struct {
	TimeFoo       time.Time `form:"time_foo" time_format:"2006-01-02" time_utc:"1" time_location:"Asia/Chongqing"`
	TimeBar       time.Time `form:"time_bar" time_format:"2006-01-02" time_utc:"1"`
	CreateTime    time.Time `form:"createTime" time_format:"unixNano"`
	UnixTime      time.Time `form:"unixTime" time_format:"unix"`
	UnixMilliTime time.Time `form:"unixMilliTime" time_format:"unixmilli"`
	UnixMicroTime time.Time `form:"unixMicroTime" time_format:"uNiXmiCrO"`
}

type FooStructForTimeTypeNotUnixFormat struct {
	CreateTime    time.Time `form:"createTime" time_format:"unixNano"`
	UnixTime      time.Time `form:"unixTime" time_format:"unix"`
	UnixMilliTime time.Time `form:"unixMilliTime" time_format:"unixMilli"`
	UnixMicroTime time.Time `form:"unixMicroTime" time_format:"unixMicro"`
}

type FooStructForTimeTypeNotFormat struct {
	TimeFoo time.Time `form:"time_foo"`
}

type FooStructForTimeTypeFailFormat struct {
	TimeFoo time.Time `form:"time_foo" time_format:"2017-11-15"`
}

type FooStructForTimeTypeFailLocation struct {
	TimeFoo time.Time `form:"time_foo" time_format:"2006-01-02" time_location:"/asia/chongqing"`
}

type FooStructForMapType struct {
	MapFoo map[string]any `form:"map_foo"`
}

type FooStructForIgnoreFormTag struct {
	Foo *string `form:"-"`
}

type InvalidNameType struct {
	TestName string `invalid_name:"test_name"`
}

type InvalidNameMapType struct {
	TestName struct {
		MapFoo map[string]any `form:"map_foo"`
	}
}

type FooStructForSliceType struct {
	SliceFoo []int `form:"slice_foo"`
}

type FooStructForStructType struct {
	StructFoo struct {
		Idx int `form:"idx"`
	}
}

type FooStructForStructPointerType struct {
	StructPointerFoo *struct {
		Name string `form:"name"`
	}
}

type FooStructForSliceMapType struct {
	// Unknown type: not support map
	SliceMapFoo []map[string]any `form:"slice_map_foo"`
}

type FooStructForBoolType struct {
	BoolFoo bool `form:"bool_foo"`
}

type FooStructForStringPtrType struct {
	PtrFoo *string `form:"ptr_foo"`
	PtrBar *string `form:"ptr_bar" binding:"required"`
}

type FooStructForMapPtrType struct {
	PtrBar *map[string]any `form:"ptr_bar"`
}

func TestBindingDefault(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingDefault(t=%v)", t)
	assert.Equal(t, Form, Default(http.MethodGet, ""))
	assert.Equal(t, Form, Default(http.MethodGet, MIMEJSON))

	assert.Equal(t, JSON, Default(http.MethodPost, MIMEJSON))
	assert.Equal(t, JSON, Default(http.MethodPut, MIMEJSON))

	assert.Equal(t, XML, Default(http.MethodPost, MIMEXML))
	assert.Equal(t, XML, Default(http.MethodPut, MIMEXML2))

	assert.Equal(t, Form, Default(http.MethodPost, MIMEPOSTForm))
	assert.Equal(t, Form, Default(http.MethodPut, MIMEPOSTForm))

	assert.Equal(t, FormMultipart, Default(http.MethodPost, MIMEMultipartPOSTForm))
	assert.Equal(t, FormMultipart, Default(http.MethodPut, MIMEMultipartPOSTForm))

	assert.Equal(t, ProtoBuf, Default(http.MethodPost, MIMEPROTOBUF))
	assert.Equal(t, ProtoBuf, Default(http.MethodPut, MIMEPROTOBUF))

	assert.Equal(t, YAML, Default(http.MethodPost, MIMEYAML))
	assert.Equal(t, YAML, Default(http.MethodPut, MIMEYAML))
	assert.Equal(t, YAML, Default(http.MethodPost, MIMEYAML2))
	assert.Equal(t, YAML, Default(http.MethodPut, MIMEYAML2))

	assert.Equal(t, TOML, Default(http.MethodPost, MIMETOML))
	assert.Equal(t, TOML, Default(http.MethodPut, MIMETOML))

	assert.Equal(t, BSON, Default(http.MethodPost, MIMEBSON))
	assert.Equal(t, BSON, Default(http.MethodPut, MIMEBSON))
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingDefault(t=%v)", t)
}

func TestBindingJSONNilBody(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONNilBody(t=%v)", t)
	__logStart := logritetime.Now()
	var obj FooStruct
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[ERROR] TestBindingJSONNilBody(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONNilBody(t=%v)", t)
}

func TestBindingJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSON(t=%v)", t)
	testBodyBinding(t,
		JSON, "json",
		"/", "/",
		`{"foo": "bar"}`, `{"bar": "foo"}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSON(t=%v)", t)
}

func TestBindingJSONSlice(t *testing.T) {
	EnableDecoderDisallowUnknownFields = true
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONSlice(t=%v)", t)
	logritelog.Default().Printf("[DEBUG] TestBindingJSONSlice(t=%v): EnableDecoderDisallowUnknownFields → %v", t, EnableDecoderDisallowUnknownFields)
	defer func() {
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONSlice(t=%v)", t)
		EnableDecoderDisallowUnknownFields = false
	}()

	testBodyBindingSlice(t, JSON, "json", "/", "/", `[]`, ``)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": ""}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": 123}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"bar": 123}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": "123456789012345678901234567890123"}]`)
}

func TestBindingJSONUseNumber(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONUseNumber(t=%v)", t)
	testBodyBindingUseNumber(t,
		JSON, "json",
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONUseNumber(t=%v)", t)
}

func TestBindingJSONUseNumber2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONUseNumber2(t=%v)", t)
	testBodyBindingUseNumber2(t,
		JSON, "json",
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONUseNumber2(t=%v)", t)
}

func TestBindingJSONDisallowUnknownFields(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONDisallowUnknownFields(t=%v)", t)
	testBodyBindingDisallowUnknownFields(t, JSON,
		"/", "/",
		`{"foo": "bar"}`, `{"foo": "bar", "what": "this"}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONDisallowUnknownFields(t=%v)", t)
}

func TestBindingJSONStringMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingJSONStringMap(t=%v)", t)
	testBodyBindingStringMap(t, JSON,
		"/", "/",
		`{"foo": "bar", "hello": "world"}`, `{"num": 2}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingJSONStringMap(t=%v)", t)
}

func TestBindingForm(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingForm(t=%v)", t)
	testFormBinding(t, http.MethodPost,
		"/", "/",
		"foo=bar&bar=foo", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingForm(t=%v)", t)
}

func TestBindingForm2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingForm2(t=%v)", t)
	testFormBinding(t, http.MethodGet,
		"/?foo=bar&bar=foo", "/?bar2=foo",
		"", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingForm2(t=%v)", t)
}

func TestBindingFormEmbeddedStruct(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormEmbeddedStruct(t=%v)", t)
	testFormBindingEmbeddedStruct(t, http.MethodPost,
		"/", "/",
		"page=1&size=2&appkey=test-appkey", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormEmbeddedStruct(t=%v)", t)
}

func TestBindingFormEmbeddedStruct2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormEmbeddedStruct2(t=%v)", t)
	testFormBindingEmbeddedStruct(t, http.MethodGet,
		"/?page=1&size=2&appkey=test-appkey", "/?bar2=foo",
		"", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormEmbeddedStruct2(t=%v)", t)
}

func TestBindingFormDefaultValue(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormDefaultValue(t=%v)", t)
	testFormBindingDefaultValue(t, http.MethodPost,
		"/", "/",
		"foo=bar", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormDefaultValue(t=%v)", t)
}

func TestBindingFormDefaultValue2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormDefaultValue2(t=%v)", t)
	testFormBindingDefaultValue(t, http.MethodGet,
		"/?foo=bar", "/?bar2=foo",
		"", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormDefaultValue2(t=%v)", t)
}

func TestBindingFormForTime(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormForTime(t=%v)", t)
	testFormBindingForTime(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15&time_bar=&createTime=1562400033000000123&unixTime=1562400033&unixMilliTime=1562400033001&unixMicroTime=1562400033000012", "bar2=foo")
	testFormBindingForTimeNotUnixFormat(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15&createTime=bad&unixTime=bad&unixMilliTime=bad&unixMicroTime=bad", "bar2=foo")
	testFormBindingForTimeNotFormat(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15", "bar2=foo")
	testFormBindingForTimeFailFormat(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15", "bar2=foo")
	testFormBindingForTimeFailLocation(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormForTime(t=%v)", t)
}

func TestBindingFormForTime2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormForTime2(t=%v)", t)
	testFormBindingForTime(t, http.MethodGet,
		"/?time_foo=2017-11-15&time_bar=&createTime=1562400033000000123&unixTime=1562400033&unixMilliTime=1562400033001&unixMicroTime=1562400033000012", "/?bar2=foo",
		"", "")
	testFormBindingForTimeNotUnixFormat(t, http.MethodPost,
		"/", "/",
		"time_foo=2017-11-15&createTime=bad&unixTime=bad&unixMilliTime=bad&unixMicroTime=bad", "bar2=foo")
	testFormBindingForTimeNotFormat(t, http.MethodGet,
		"/?time_foo=2017-11-15", "/?bar2=foo",
		"", "")
	testFormBindingForTimeFailFormat(t, http.MethodGet,
		"/?time_foo=2017-11-15", "/?bar2=foo",
		"", "")
	testFormBindingForTimeFailLocation(t, http.MethodGet,
		"/?time_foo=2017-11-15", "/?bar2=foo",
		"", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormForTime2(t=%v)", t)
}

func TestFormBindingIgnoreField(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormBindingIgnoreField(t=%v)", t)
	testFormBindingIgnoreField(t, http.MethodPost,
		"/", "/",
		"-=bar", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestFormBindingIgnoreField(t=%v)", t)
}

func TestBindingFormInvalidName(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormInvalidName(t=%v)", t)
	testFormBindingInvalidName(t, http.MethodPost,
		"/", "/",
		"test_name=bar", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormInvalidName(t=%v)", t)
}

func TestBindingFormInvalidName2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormInvalidName2(t=%v)", t)
	testFormBindingInvalidName2(t, http.MethodPost,
		"/", "/",
		"map_foo=bar", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormInvalidName2(t=%v)", t)
}

func TestBindingFormForType(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormForType(t=%v)", t)
	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"map_foo={\"bar\":123}", "map_foo=1", "Map")

	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"slice_foo=1&slice_foo=2", "bar2=1&bar2=2", "Slice")

	testFormBindingForType(t, http.MethodGet,
		"/?slice_foo=1&slice_foo=2", "/?bar2=1&bar2=2",
		"", "", "Slice")

	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"slice_map_foo=1&slice_map_foo=2", "bar2=1&bar2=2", "SliceMap")

	testFormBindingForType(t, http.MethodGet,
		"/?slice_map_foo=1&slice_map_foo=2", "/?bar2=1&bar2=2",
		"", "", "SliceMap")

	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"ptr_bar=test", "bar2=test", "Ptr")

	testFormBindingForType(t, http.MethodGet,
		"/?ptr_bar=test", "/?bar2=test",
		"", "", "Ptr")

	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"idx=123", "id1=1", "Struct")

	testFormBindingForType(t, http.MethodGet,
		"/?idx=123", "/?id1=1",
		"", "", "Struct")

	testFormBindingForType(t, http.MethodPost,
		"/", "/",
		"name=thinkerou", "name1=ou", "StructPointer")

	testFormBindingForType(t, http.MethodGet,
		"/?name=thinkerou", "/?name1=ou",
		"", "", "StructPointer")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormForType(t=%v)", t)
}

func TestBindingFormStringMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormStringMap(t=%v)", t)
	testBodyBindingStringMap(t, Form,
		"/", "",
		`foo=bar&hello=world`, "")
	// Should pick the last value
	testBodyBindingStringMap(t, Form,
		"/", "",
		`foo=something&foo=bar&hello=world`, "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormStringMap(t=%v)", t)
}

func TestBindingFormStringSliceMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormStringSliceMap(t=%v)", t)
	obj := make(map[string][]string)
	req := requestWithBody(http.MethodPost, "/", "foo=something&foo=bar&hello=world")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err := Form.Bind(req, &obj)
	require.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	target := map[string][]string{
		"foo":   {"something", "bar"},
		"hello": {"world"},
	}
	assert.True(t, reflect.DeepEqual(obj, target))

	objInvalid := make(map[string][]int)
	req = requestWithBody(http.MethodPost, "/", "foo=something&foo=bar&hello=world")
	req.Header.Add("Content-Type", MIMEPOSTForm)
	err = Form.Bind(req, &objInvalid)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] TestBindingFormStringSliceMap(t=%v): req → %v", t, req)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormStringSliceMap(t=%v)", t)
	logritelog.Default().Printf("[DEBUG] TestBindingFormStringSliceMap(t=%v): err → %v", t, err)
}

func TestBindingQuery(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQuery(t=%v)", t)
	testQueryBinding(t, http.MethodPost,
		"/?foo=bar&bar=foo", "/",
		"foo=unused", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQuery(t=%v)", t)
}

func TestBindingQuery2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQuery2(t=%v)", t)
	testQueryBinding(t, http.MethodGet,
		"/?foo=bar&bar=foo", "/?bar2=foo",
		"foo=unused", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQuery2(t=%v)", t)
}

func TestBindingQueryFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQueryFail(t=%v)", t)
	testQueryBindingFail(t, http.MethodPost,
		"/?map_foo=", "/",
		"map_foo=unused", "bar2=foo")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQueryFail(t=%v)", t)
}

func TestBindingQueryFail2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQueryFail2(t=%v)", t)
	testQueryBindingFail(t, http.MethodGet,
		"/?map_foo=", "/?bar2=foo",
		"map_foo=unused", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQueryFail2(t=%v)", t)
}

func TestBindingQueryBoolFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQueryBoolFail(t=%v)", t)
	testQueryBindingBoolFail(t, http.MethodGet,
		"/?bool_foo=fasl", "/?bar2=foo",
		"bool_foo=unused", "")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQueryBoolFail(t=%v)", t)
}

func TestBindingQueryStringMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingQueryStringMap(t=%v)", t)
	b := Query

	obj := make(map[string]string)
	req := requestWithBody(http.MethodGet, "/?foo=bar&hello=world", "")
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "bar", obj["foo"])
	assert.Equal(t, "world", obj["hello"])

	obj = make(map[string]string)
	req = requestWithBody(http.MethodGet, "/?foo=bar&foo=2&hello=world", "") // should pick last
	logritelog.Default().Printf("[DEBUG] TestBindingQueryStringMap(t=%v): req → %v", t, req)
	logritelog.Default().Printf("[DEBUG] TestBindingQueryStringMap(t=%v): obj → %v", t, obj)
	err = b.Bind(req, &obj)
	require.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "2", obj["foo"])
	assert.Equal(t, "world", obj["hello"])
	logritelog.Default().Printf("[DEBUG] TestBindingQueryStringMap(t=%v): err → %v", t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingQueryStringMap(t=%v)", t)
}

func TestBindingXML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingXML(t=%v)", t)
	testBodyBinding(t,
		XML, "xml",
		"/", "/",
		"<map><foo>bar</foo></map>", "<map><bar>foo</bar></map>")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingXML(t=%v)", t)
}

func TestBindingXMLFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingXMLFail(t=%v)", t)
	testBodyBindingFail(t,
		XML, "xml",
		"/", "/",
		"<map><foo>bar<foo></map>", "<map><bar>foo</bar></map>")
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingXMLFail(t=%v)", t)
}

func TestBindingTOML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingTOML(t=%v)", t)
	testBodyBinding(t,
		TOML, "toml",
		"/", "/",
		`foo="bar"`, `bar="foo"`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingTOML(t=%v)", t)
}

func TestBindingTOMLFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingTOMLFail(t=%v)", t)
	testBodyBindingFail(t,
		TOML, "toml",
		"/", "/",
		`foo=\n"bar"`, `bar="foo"`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingTOMLFail(t=%v)", t)
}

func TestBindingYAML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingYAML(t=%v)", t)
	testBodyBinding(t,
		YAML, "yaml",
		"/", "/",
		`foo: bar`, `bar: foo`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingYAML(t=%v)", t)
}

func TestBindingYAMLStringMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingYAMLStringMap(t=%v)", t)
	// YAML is a superset of JSON, so the test below is JSON (to avoid newlines)
	testBodyBindingStringMap(t, YAML,
		"/", "/",
		`{"foo": "bar", "hello": "world"}`, `{"nested": {"foo": "bar"}}`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingYAMLStringMap(t=%v)", t)
}

func TestBindingYAMLFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingYAMLFail(t=%v)", t)
	testBodyBindingFail(t,
		YAML, "yaml",
		"/", "/",
		`foo:\nbar`, `bar: foo`)
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingYAMLFail(t=%v)", t)
}

func createFormPostRequest(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormPostRequest(t=%v)", t)
	req, err := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", bytes.NewBufferString("foo=bar&bar=foo"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEPOSTForm)
	logritelog.Default().Printf("[ERROR] createFormPostRequest(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormPostRequest(t=%v)", t)
	return req
}

func createDefaultFormPostRequest(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createDefaultFormPostRequest(t=%v)", t)
	req, err := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", bytes.NewBufferString("foo=bar"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEPOSTForm)
	logritelog.Default().Printf("[ERROR] createDefaultFormPostRequest(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createDefaultFormPostRequest(t=%v)", t)
	return req
}

func createFormPostRequestForMap(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormPostRequestForMap(t=%v)", t)
	req, err := http.NewRequest(http.MethodPost, "/?map_foo=getfoo", bytes.NewBufferString("map_foo={\"bar\":123}"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEPOSTForm)
	logritelog.Default().Printf("[ERROR] createFormPostRequestForMap(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormPostRequestForMap(t=%v)", t)
	return req
}

func createFormPostRequestForMapFail(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormPostRequestForMapFail(t=%v)", t)
	req, err := http.NewRequest(http.MethodPost, "/?map_foo=getfoo", bytes.NewBufferString("map_foo=hello"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEPOSTForm)
	logritelog.Default().Printf("[ERROR] createFormPostRequestForMapFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormPostRequestForMapFail(t=%v)", t)
	return req
}

func createFormFilesMultipartRequest(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormFilesMultipartRequest(t=%v)", t)
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	require.NoError(t, mw.SetBoundary(boundary))
	require.NoError(t, mw.WriteField("foo", "bar"))
	require.NoError(t, mw.WriteField("bar", "foo"))

	f, err := os.Open("form.go")
	require.NoError(t, err)
	defer f.Close()
	fw, err1 := mw.CreateFormFile("file", "form.go")
	require.NoError(t, err1)
	_, err = io.Copy(fw, f)
	require.NoError(t, err)

	req, err2 := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", body)
	require.NoError(t, err2)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)

	logritelog.Default().Printf("[ERROR] createFormFilesMultipartRequest(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormFilesMultipartRequest(t=%v)", t)
	return req
}

func createFormFilesMultipartRequestFail(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormFilesMultipartRequestFail(t=%v)", t)
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	require.NoError(t, mw.SetBoundary(boundary))
	require.NoError(t, mw.WriteField("foo", "bar"))
	require.NoError(t, mw.WriteField("bar", "foo"))

	f, err := os.Open("form.go")
	require.NoError(t, err)
	defer f.Close()
	fw, err1 := mw.CreateFormFile("file_foo", "form_foo.go")
	require.NoError(t, err1)
	_, err = io.Copy(fw, f)
	require.NoError(t, err)

	req, err2 := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", body)
	require.NoError(t, err2)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)

	logritelog.Default().Printf("[ERROR] createFormFilesMultipartRequestFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormFilesMultipartRequestFail(t=%v)", t)
	return req
}

func createFormMultipartRequest(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormMultipartRequest(t=%v)", t)
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	require.NoError(t, mw.SetBoundary(boundary))
	require.NoError(t, mw.WriteField("foo", "bar"))
	require.NoError(t, mw.WriteField("bar", "foo"))
	req, err := http.NewRequest(http.MethodPost, "/?foo=getfoo&bar=getbar", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)
	logritelog.Default().Printf("[ERROR] createFormMultipartRequest(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormMultipartRequest(t=%v)", t)
	return req
}

func createFormMultipartRequestForMap(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormMultipartRequestForMap(t=%v)", t)
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	require.NoError(t, mw.SetBoundary(boundary))
	require.NoError(t, mw.WriteField("map_foo", "{\"bar\":123, \"name\":\"thinkerou\", \"pai\": 3.14}"))
	req, err := http.NewRequest(http.MethodPost, "/?map_foo=getfoo", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)
	logritelog.Default().Printf("[ERROR] createFormMultipartRequestForMap(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormMultipartRequestForMap(t=%v)", t)
	return req
}

func createFormMultipartRequestForMapFail(t *testing.T) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createFormMultipartRequestForMapFail(t=%v)", t)
	boundary := "--testboundary"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	defer mw.Close()

	require.NoError(t, mw.SetBoundary(boundary))
	require.NoError(t, mw.WriteField("map_foo", "3.14"))
	req, err := http.NewRequest(http.MethodPost, "/?map_foo=getfoo", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+boundary)
	logritelog.Default().Printf("[ERROR] createFormMultipartRequestForMapFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createFormMultipartRequestForMapFail(t=%v)", t)
	return req
}

func TestBindingFormPost(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormPost(t=%v)", t)
	req := createFormPostRequest(t)
	var obj FooBarStruct
	require.NoError(t, FormPost.Bind(req, &obj))

	assert.Equal(t, "form-urlencoded", FormPost.Name())
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormPost(t=%v)", t)
}

func TestBindingDefaultValueFormPost(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingDefaultValueFormPost(t=%v)", t)
	req := createDefaultFormPostRequest(t)
	var obj FooDefaultBarStruct
	require.NoError(t, FormPost.Bind(req, &obj))

	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "hello", obj.Bar)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingDefaultValueFormPost(t=%v)", t)
}

func TestBindingFormPostForMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormPostForMap(t=%v)", t)
	req := createFormPostRequestForMap(t)
	var obj FooStructForMapType
	err := FormPost.Bind(req, &obj)
	require.NoError(t, err)
	assert.InDelta(t, float64(123), obj.MapFoo["bar"].(float64), 0.01)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormPostForMap(t=%v)", t)
}

func TestBindingFormPostForMapFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormPostForMapFail(t=%v)", t)
	req := createFormPostRequestForMapFail(t)
	var obj FooStructForMapType
	err := FormPost.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormPostForMapFail(t=%v)", t)
}

func TestBindingFormFilesMultipart(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormFilesMultipart(t=%v)", t)
	req := createFormFilesMultipartRequest(t)
	var obj FooBarFileStruct
	err := FormMultipart.Bind(req, &obj)
	require.NoError(t, err)

	// file from os
	f, _ := os.Open("form.go")
	defer f.Close()
	fileActual, _ := io.ReadAll(f)

	// file from multipart
	mf, _ := obj.File.Open()
	defer mf.Close()
	fileExpect, _ := io.ReadAll(mf)

	assert.Equal(t, "multipart/form-data", FormMultipart.Name())
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, fileExpect, fileActual)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormFilesMultipart(t=%v)", t)
}

func TestBindingFormFilesMultipartFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormFilesMultipartFail(t=%v)", t)
	req := createFormFilesMultipartRequestFail(t)
	var obj FooBarFileFailStruct
	err := FormMultipart.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormFilesMultipartFail(t=%v)", t)
}

func TestBindingFormMultipart(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormMultipart(t=%v)", t)
	req := createFormMultipartRequest(t)
	var obj FooBarStruct
	require.NoError(t, FormMultipart.Bind(req, &obj))

	assert.Equal(t, "multipart/form-data", FormMultipart.Name())
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormMultipart(t=%v)", t)
}

func TestBindingFormMultipartForMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormMultipartForMap(t=%v)", t)
	req := createFormMultipartRequestForMap(t)
	var obj FooStructForMapType
	err := FormMultipart.Bind(req, &obj)
	require.NoError(t, err)
	assert.InDelta(t, float64(123), obj.MapFoo["bar"].(float64), 0.01)
	assert.Equal(t, "thinkerou", obj.MapFoo["name"].(string))
	assert.InDelta(t, float64(3.14), obj.MapFoo["pai"].(float64), 0.01)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormMultipartForMap(t=%v)", t)
}

func TestBindingFormMultipartForMapFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingFormMultipartForMapFail(t=%v)", t)
	req := createFormMultipartRequestForMapFail(t)
	var obj FooStructForMapType
	err := FormMultipart.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindingFormMultipartForMapFail(t=%v)", t)
}

func TestBindingProtoBuf(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingProtoBuf(t=%v)", t)
	test := &protoexample.Test{
		Label: proto.String("yes"),
	}
	data, _ := proto.Marshal(test)

	testProtoBodyBinding(t,
		ProtoBuf, "protobuf",
		"/", "/",
		string(data), string(data[1:]))
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingProtoBuf(t=%v)", t)
}

func TestBindingProtoBufFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingProtoBufFail(t=%v)", t)
	test := &protoexample.Test{
		Label: proto.String("yes"),
	}
	data, _ := proto.Marshal(test)

	testProtoBodyBindingFail(t,
		ProtoBuf, "protobuf",
		"/", "/",
		string(data), string(data[1:]))
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingProtoBufFail(t=%v)", t)
}

func TestBindingBSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindingBSON(t=%v)", t)
	var obj FooStruct
	obj.Foo = "bar"
	logritelog.Default().Printf("[DEBUG] TestBindingBSON(t=%v): obj.Foo → %v", t, obj.Foo)
	data, _ := bson.Marshal(&obj)
	testBodyBinding(t,
		BSON, "bson",
		"/", "/",
		string(data),
		// note: for badbody, we remove first byte to make it invalid
		string(data[1:]))
		logritelog.Default().Printf("[WARN] <<< Exiting TestBindingBSON(t=%v)", t)
}

func TestValidationFails(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestValidationFails(t=%v)", t)
	var obj FooStruct
	req := requestWithBody(http.MethodPost, "/", `{"bar": "foo"}`)
	err := JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestValidationFails(t=%v)", t)
}

func TestValidationDisabled(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestValidationDisabled(t=%v)", t)
	backup := Validator
	Validator = nil
	logritelog.Default().Printf("[DEBUG] TestValidationDisabled(t=%v): Validator → %v", t, Validator)
	logritelog.Default().Printf("[WARN] <<< Exiting TestValidationDisabled(t=%v)", t)
	defer func() { Validator = backup }()

	var obj FooStruct
	req := requestWithBody(http.MethodPost, "/", `{"bar": "foo"}`)
	err := JSON.Bind(req, &obj)
	require.NoError(t, err)
}

func TestRequiredSucceeds(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRequiredSucceeds(t=%v)", t)
	type HogeStruct struct {
		Hoge *int `json:"hoge" binding:"required"`
	}

	var obj HogeStruct
	req := requestWithBody(http.MethodPost, "/", `{"hoge": 0}`)
	err := JSON.Bind(req, &obj)
	require.NoError(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRequiredSucceeds(t=%v)", t)
}

func TestRequiredFails(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRequiredFails(t=%v)", t)
	type HogeStruct struct {
		Hoge *int `json:"foo" binding:"required"`
	}

	var obj HogeStruct
	req := requestWithBody(http.MethodPost, "/", `{"boen": 0}`)
	err := JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRequiredFails(t=%v)", t)
}

func TestHeaderBinding(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestHeaderBinding(t=%v)", t)
	h := Header
	assert.Equal(t, "header", h.Name())

	type tHeader struct {
		Limit int `header:"limit"`
	}

	var theader tHeader
	req := requestWithBody(http.MethodGet, "/", "")
	req.Header.Add("limit", "1000")
	require.NoError(t, h.Bind(req, &theader))
	assert.Equal(t, 1000, theader.Limit)

	req = requestWithBody(http.MethodGet, "/", "")
	req.Header.Add("fail", `{fail:fail}`)

	logritelog.Default().Printf("[DEBUG] TestHeaderBinding(t=%v): req → %v", t, req)
	type failStruct struct {
		Fail map[string]any `header:"fail"`
	}

	err := h.Bind(req, &failStruct{})
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestHeaderBinding(t=%v)", t)
}

func TestUriBinding(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestUriBinding(t=%v)", t)
	b := Uri
	assert.Equal(t, "uri", b.Name())

	type Tag struct {
		Name string `uri:"name"`
	}
	var tag Tag
	m := make(map[string][]string)
	m["name"] = []string{"thinkerou"}
	require.NoError(t, b.BindUri(m, &tag))
	assert.Equal(t, "thinkerou", tag.Name)

	type NotSupportStruct struct {
		Name map[string]any `uri:"name"`
	}
	var not NotSupportStruct
	require.Error(t, b.BindUri(m, &not))
	assert.Equal(t, map[string]any(nil), not.Name)
	logritelog.Default().Printf("[WARN] <<< Exiting TestUriBinding(t=%v)", t)
}

func TestUriInnerBinding(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestUriInnerBinding(t=%v)", t)
	type Tag struct {
		Name string `uri:"name"`
		S    struct {
			Age int `uri:"age"`
		}
	}

	expectedName := "mike"
	expectedAge := 25

	m := map[string][]string{
		"name": {expectedName},
		"age":  {strconv.Itoa(expectedAge)},
	}

	var tag Tag
	require.NoError(t, Uri.BindUri(m, &tag))
	assert.Equal(t, expectedName, tag.Name)
	assert.Equal(t, expectedAge, tag.S.Age)
	logritelog.Default().Printf("[WARN] <<< Exiting TestUriInnerBinding(t=%v)", t)
}

func testFormBindingEmbeddedStruct(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingEmbeddedStruct(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := QueryTest{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, 1, obj.Page)
	assert.Equal(t, 2, obj.Size)
	assert.Equal(t, "test-appkey", obj.Appkey)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingEmbeddedStruct(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
}

func testFormBinding(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooBarStruct{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)

	obj = FooBarStruct{}
	logritelog.Default().Printf("[DEBUG] testFormBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[DEBUG] testFormBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingDefaultValue(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingDefaultValue(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooDefaultBarStruct{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "hello", obj.Bar)

	obj = FooDefaultBarStruct{}
	logritelog.Default().Printf("[DEBUG] testFormBindingDefaultValue(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingDefaultValue(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingDefaultValue(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingDefaultValue(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func TestFormBindingFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormBindingFail(t=%v)", t)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooBarStruct{}
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := b.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[ERROR] TestFormBindingFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestFormBindingFail(t=%v)", t)
}

func TestFormBindingMultipartFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormBindingMultipartFail(t=%v)", t)
	obj := FooBarStruct{}
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("foo=bar"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+";boundary=testboundary")
	_, err = req.MultipartReader()
	require.NoError(t, err)
	err = Form.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] TestFormBindingMultipartFail(t=%v): err → %v", t, err)
	logritelog.Default().Printf("[ERROR] TestFormBindingMultipartFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestFormBindingMultipartFail(t=%v)", t)
}

func TestFormPostBindingFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormPostBindingFail(t=%v)", t)
	b := FormPost
	assert.Equal(t, "form-urlencoded", b.Name())

	obj := FooBarStruct{}
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := b.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[ERROR] TestFormPostBindingFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestFormPostBindingFail(t=%v)", t)
}

func TestFormMultipartBindingFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormMultipartBindingFail(t=%v)", t)
	b := FormMultipart
	assert.Equal(t, "multipart/form-data", b.Name())

	obj := FooBarStruct{}
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := b.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[ERROR] TestFormMultipartBindingFail(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestFormMultipartBindingFail(t=%v)", t)
}

func testFormBindingForTime(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForTime(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooBarStructForTimeType{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)

	require.NoError(t, err)
	assert.Equal(t, int64(1510675200), obj.TimeFoo.Unix())
	assert.Equal(t, "Asia/Chongqing", obj.TimeFoo.Location().String())
	assert.Equal(t, int64(-62135596800), obj.TimeBar.Unix())
	assert.Equal(t, "UTC", obj.TimeBar.Location().String())
	assert.Equal(t, int64(1562400033000000123), obj.CreateTime.UnixNano())
	assert.Equal(t, int64(1562400033), obj.UnixTime.Unix())
	assert.Equal(t, int64(1562400033001), obj.UnixMilliTime.UnixMilli())
	assert.Equal(t, int64(1562400033000012), obj.UnixMicroTime.UnixMicro())

	obj = FooBarStructForTimeType{}
	logritelog.Default().Printf("[DEBUG] testFormBindingForTime(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTime(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForTime(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTime(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingForTimeNotUnixFormat(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForTimeNotUnixFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooStructForTimeTypeNotUnixFormat{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)

	obj = FooStructForTimeTypeNotUnixFormat{}
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotUnixFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotUnixFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForTimeNotUnixFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotUnixFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingForTimeNotFormat(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForTimeNotFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooStructForTimeTypeNotFormat{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)

	obj = FooStructForTimeTypeNotFormat{}
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForTimeNotFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeNotFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingForTimeFailFormat(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForTimeFailFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooStructForTimeTypeFailFormat{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)

	obj = FooStructForTimeTypeFailFormat{}
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForTimeFailFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailFormat(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingForTimeFailLocation(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForTimeFailLocation(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooStructForTimeTypeFailLocation{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)

	obj = FooStructForTimeTypeFailLocation{}
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailLocation(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailLocation(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForTimeFailLocation(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingForTimeFailLocation(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingIgnoreField(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingIgnoreField(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := FooStructForIgnoreFormTag{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)

	assert.Nil(t, obj.Foo)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingIgnoreField(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
}

func testFormBindingInvalidName(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingInvalidName(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := InvalidNameType{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Empty(t, obj.TestName)

	obj = InvalidNameType{}
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingInvalidName(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingInvalidName2(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingInvalidName2(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Form
	assert.Equal(t, "form", b.Name())

	obj := InvalidNameMapType{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)

	obj = InvalidNameMapType{}
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName2(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, method, path, badPath, body, badBody, obj)
	req = requestWithBody(method, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName2(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, method, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingInvalidName2(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testFormBindingInvalidName2(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, method, path, badPath, body, badBody, err)
}

func testFormBindingForType(t *testing.T, method, path, badPath, body, badBody string, typ string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testFormBindingForType(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v,typ=%v)", t, method, path, badPath, body, badBody, typ)
	b := Form
	assert.Equal(t, "form", b.Name())

	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	switch typ {
	case "Slice":
		obj := FooStructForSliceType{}
		err := b.Bind(req, &obj)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, obj.SliceFoo)

		obj = FooStructForSliceType{}
		logritelog.Default().Printf("[DEBUG] testFormBindingForType(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v,typ=%v): obj → %v", t, method, path, badPath, body, badBody, typ, obj)
		req = requestWithBody(method, badPath, badBody)
		err = JSON.Bind(req, &obj)
		require.Error(t, err)
	case "Struct":
		logritelog.Default().Printf("[DEBUG] testFormBindingForType(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v,typ=%v): req → %v", t, method, path, badPath, body, badBody, typ, req)
		obj := FooStructForStructType{}
		err := b.Bind(req, &obj)
		logritelog.Default().Printf("[DEBUG] testFormBindingForType(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v,typ=%v): err → %v", t, method, path, badPath, body, badBody, typ, err)
		require.NoError(t, err)
		assert.Equal(t,
			struct {
				Idx int "form:\"idx\""
			}{Idx: 123},
			obj.StructFoo)
	case "StructPointer":
		obj := FooStructForStructPointerType{}
		err := b.Bind(req, &obj)
		require.NoError(t, err)
		assert.Equal(t,
			struct {
				Name string "form:\"name\""
			}{Name: "thinkerou"},
			*obj.StructPointerFoo)
	case "Map":
		obj := FooStructForMapType{}
		err := b.Bind(req, &obj)
		require.NoError(t, err)
		assert.InDelta(t, float64(123), obj.MapFoo["bar"].(float64), 0.01)
	case "SliceMap":
		obj := FooStructForSliceMapType{}
		err := b.Bind(req, &obj)
		require.Error(t, err)
	case "Ptr":
		obj := FooStructForStringPtrType{}
		err := b.Bind(req, &obj)
		require.NoError(t, err)
		assert.Nil(t, obj.PtrFoo)
		assert.Equal(t, "test", *obj.PtrBar)

		obj = FooStructForStringPtrType{}
		obj.PtrBar = new(string)
		err = b.Bind(req, &obj)
		require.NoError(t, err)
		assert.Equal(t, "test", *obj.PtrBar)

		objErr := FooStructForMapPtrType{}
		err = b.Bind(req, &objErr)
		require.Error(t, err)

		obj = FooStructForStringPtrType{}
		req = requestWithBody(method, badPath, badBody)
		err = b.Bind(req, &obj)
		require.Error(t, err)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting testFormBindingForType(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v,typ=%v)", t, method, path, badPath, body, badBody, typ)
}

func testQueryBinding(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testQueryBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Query
	assert.Equal(t, "query", b.Name())

	obj := FooBarStruct{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, "foo", obj.Bar)
	logritelog.Default().Printf("[WARN] <<< Exiting testQueryBinding(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
}

func testQueryBindingFail(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testQueryBindingFail(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Query
	assert.Equal(t, "query", b.Name())

	obj := FooStructForMapType{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting testQueryBindingFail(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
}

func testQueryBindingBoolFail(t *testing.T, method, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testQueryBindingBoolFail(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
	b := Query
	assert.Equal(t, "query", b.Name())

	obj := FooStructForBoolType{}
	req := requestWithBody(method, path, body)
	if method == http.MethodPost {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting testQueryBindingBoolFail(t=%v,method=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, method, path, badPath, body, badBody)
}

func testBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := FooStruct{}
	req := requestWithBody(http.MethodPost, path, body)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStruct{}
	logritelog.Default().Printf("[DEBUG] testBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

func testBodyBindingSlice(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingSlice(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	var obj1 []FooStruct
	req := requestWithBody(http.MethodPost, path, body)
	err := b.Bind(req, &obj1)
	require.NoError(t, err)

	var obj2 []FooStruct
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj2)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testBodyBindingSlice(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingSlice(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBindingSlice(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

func testBodyBindingStringMap(t *testing.T, b Binding, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingStringMap(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, path, badPath, body, badBody)
	obj := make(map[string]string)
	req := requestWithBody(http.MethodPost, path, body)
	if b.Name() == "form" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "bar", obj["foo"])
	assert.Equal(t, "world", obj["hello"])

	if badPath != "" && badBody != "" {
		obj = make(map[string]string)
		req = requestWithBody(http.MethodPost, badPath, badBody)
		err = b.Bind(req, &obj)
		require.Error(t, err)
	}
	logritelog.Default().Printf("[DEBUG] testBodyBindingStringMap(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, path, badPath, body, badBody, obj)

	objInt := make(map[string]int)
	req = requestWithBody(http.MethodPost, path, body)
	logritelog.Default().Printf("[DEBUG] testBodyBindingStringMap(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, path, badPath, body, badBody, req)
	err = b.Bind(req, &objInt)
	logritelog.Default().Printf("[DEBUG] testBodyBindingStringMap(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, path, badPath, body, badBody, err)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingStringMap(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, path, badPath, body, badBody)
}

func testBodyBindingUseNumber(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := FooStructUseNumber{}
	req := requestWithBody(http.MethodPost, path, body)
	EnableDecoderUseNumber = true
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): EnableDecoderUseNumber → %v", t, b, name, path, badPath, body, badBody, EnableDecoderUseNumber)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	// we hope it is int64(123)
	v, e := obj.Foo.(json.Number).Int64()
	require.NoError(t, e)
	assert.Equal(t, int64(123), v)

	obj = FooStructUseNumber{}
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

func testBodyBindingUseNumber2(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := FooStructUseNumber{}
	req := requestWithBody(http.MethodPost, path, body)
	EnableDecoderUseNumber = false
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): EnableDecoderUseNumber → %v", t, b, name, path, badPath, body, badBody, EnableDecoderUseNumber)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	// it will return float64(123) if not use EnableDecoderUseNumber
	// maybe it is not hoped
	assert.InDelta(t, float64(123), obj.Foo, 0.01)

	obj = FooStructUseNumber{}
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBindingUseNumber2(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

func testBodyBindingDisallowUnknownFields(t *testing.T, b Binding, path, badPath, body, badBody string) {
	EnableDecoderDisallowUnknownFields = true
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingDisallowUnknownFields(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBindingDisallowUnknownFields(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v): EnableDecoderDisallowUnknownFields → %v", t, b, path, badPath, body, badBody, EnableDecoderDisallowUnknownFields)
	defer func() {
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingDisallowUnknownFields(t=%v,b=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, path, badPath, body, badBody)
		EnableDecoderDisallowUnknownFields = false
	}()

	obj := FooStructDisallowUnknownFields{}
	req := requestWithBody(http.MethodPost, path, body)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStructDisallowUnknownFields{}
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "what")
}

func testBodyBindingFail(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := FooStruct{}
	req := requestWithBody(http.MethodPost, path, body)
	err := b.Bind(req, &obj)
	require.Error(t, err)
	assert.Empty(t, obj.Foo)

	obj = FooStruct{}
	logritelog.Default().Printf("[DEBUG] testBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	err = JSON.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

func testProtoBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testProtoBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := protoexample.Test{}
	req := requestWithBody(http.MethodPost, path, body)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err := b.Bind(req, &obj)
	require.NoError(t, err)
	assert.Equal(t, "yes", *obj.Label)

	obj = protoexample.Test{}
	logritelog.Default().Printf("[DEBUG] testProtoBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = ProtoBuf.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testProtoBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testProtoBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	logritelog.Default().Printf("[DEBUG] testProtoBodyBinding(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
}

type hook struct{}

func (h hook) Read([]byte) (int, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Read()")
	return 0, errors.New("error")
}

type failRead struct{}

func (f *failRead) Read(b []byte) (n int, err error) {
	return 0, errors.New("my fail")
}

func (f *failRead) Close() error {
	logritelog.Default().Printf("[ERROR] >>> Entering Close()")
	logritelog.Default().Printf("[WARN] <<< Exiting Close()")
	return nil
}

func TestPlainBinding(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestPlainBinding(t=%v)", t)
	p := Plain
	assert.Equal(t, "plain", p.Name())

	var s string
	req := requestWithBody(http.MethodPost, "/", "test string")
	require.NoError(t, p.Bind(req, &s))
	assert.Equal(t, "test string", s)

	var bs []byte
	req = requestWithBody(http.MethodPost, "/", "test []byte")
	require.NoError(t, p.Bind(req, &bs))
	assert.Equal(t, bs, []byte("test []byte"))

	logritelog.Default().Printf("[DEBUG] TestPlainBinding(t=%v): req → %v", t, req)
	var i int
	req = requestWithBody(http.MethodPost, "/", "test fail")
	require.Error(t, p.Bind(req, &i))

	req = requestWithBody(http.MethodPost, "/", "")
	req.Body = &failRead{}
	logritelog.Default().Printf("[DEBUG] TestPlainBinding(t=%v): req.Body → %v", t, req.Body)
	require.Error(t, p.Bind(req, &s))

	req = requestWithBody(http.MethodPost, "/", "")
	require.NoError(t, p.Bind(req, nil))

	var ptr *string
	req = requestWithBody(http.MethodPost, "/", "")
	require.NoError(t, p.Bind(req, ptr))
	logritelog.Default().Printf("[WARN] <<< Exiting TestPlainBinding(t=%v)", t)
}

func testProtoBodyBindingFail(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	logritelog.Default().Printf("[ERROR] >>> Entering testProtoBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
	assert.Equal(t, name, b.Name())

	obj := protoexample.Test{}
	req := requestWithBody(http.MethodPost, path, body)

	req.Body = io.NopCloser(&hook{})
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err := b.Bind(req, &obj)
	require.Error(t, err)

	invalidobj := FooStruct{}
	req.Body = io.NopCloser(strings.NewReader(`{"msg":"hello"}`))
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = b.Bind(req, &invalidobj)
	require.Error(t, err)
	assert.Equal(t, "obj is not ProtoMessage", err.Error())

	logritelog.Default().Printf("[DEBUG] testProtoBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): err → %v", t, b, name, path, badPath, body, badBody, err)
	obj = protoexample.Test{}
	logritelog.Default().Printf("[DEBUG] testProtoBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): obj → %v", t, b, name, path, badPath, body, badBody, obj)
	req = requestWithBody(http.MethodPost, badPath, badBody)
	req.Header.Add("Content-Type", MIMEPROTOBUF)
	err = ProtoBuf.Bind(req, &obj)
	require.Error(t, err)
	logritelog.Default().Printf("[DEBUG] testProtoBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v): req → %v", t, b, name, path, badPath, body, badBody, req)
	logritelog.Default().Printf("[WARN] <<< Exiting testProtoBodyBindingFail(t=%v,b=%v,name=%v,path=%v,badPath=%v,body=%v,badBody=%v)", t, b, name, path, badPath, body, badBody)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	logritelog.Default().Printf("[ERROR] >>> Entering requestWithBody(method=%v,path=%v,body=%v)", method, path, body)
	logritelog.Default().Printf("[ERROR] requestWithBody(method,path,body)=%v,%v,%v: external service call took %v ms", method, path, body, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting requestWithBody(method=%v,path=%v,body=%v)", method, path, body)
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}
