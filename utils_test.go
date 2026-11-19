// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	SetMode(TestMode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

func BenchmarkParseAccept(b *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkParseAccept(b=%v)", b)
	for b.Loop() {
		parseAccept("text/html , application/xhtml+xml,application/xml;q=0.9,  */* ;q=0.8")
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkParseAccept(b=%v)", b)
}

type testStruct struct {
	T *testing.T
}

func (t *testStruct) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logritelog.Default().Printf("[ERROR] >>> Entering ServeHTTP(w=%v,req=%v)", w, req)
	assert.Equal(t.T, http.MethodPost, req.Method)
	assert.Equal(t.T, "/path", req.URL.Path)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "hello")
	logritelog.Default().Printf("[WARN] <<< Exiting ServeHTTP(w=%v,req=%v)", w, req)
}

func TestWrap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestWrap(t=%v)", t)
	router := New()
	router.POST("/path", WrapH(&testStruct{t}))
	router.GET("/path2", WrapF(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "/path2", req.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "hola!")
	}))

	w := PerformRequest(router, http.MethodPost, "/path")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "hello", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/path2")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "hola!", w.Body.String())
	logritelog.Default().Printf("[WARN] <<< Exiting TestWrap(t=%v)", t)
}

func TestLastChar(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLastChar(t=%v)", t)
	assert.Equal(t, uint8('a'), lastChar("hola"))
	assert.Equal(t, uint8('s'), lastChar("adios"))
	assert.Panics(t, func() { lastChar("") })
	logritelog.Default().Printf("[WARN] <<< Exiting TestLastChar(t=%v)", t)
}

func TestParseAccept(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestParseAccept(t=%v)", t)
	parts := parseAccept("text/html , application/xhtml+xml,application/xml;q=0.9,  */* ;q=0.8")
	assert.Len(t, parts, 4)
	assert.Equal(t, "text/html", parts[0])
	assert.Equal(t, "application/xhtml+xml", parts[1])
	assert.Equal(t, "application/xml", parts[2])
	assert.Equal(t, "*/*", parts[3])
	logritelog.Default().Printf("[WARN] <<< Exiting TestParseAccept(t=%v)", t)
}

func TestChooseData(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestChooseData(t=%v)", t)
	A := "a"
	B := "b"
	assert.Equal(t, A, chooseData(A, B))
	assert.Equal(t, B, chooseData(nil, B))
	assert.Panics(t, func() { chooseData(nil, nil) })
	logritelog.Default().Printf("[WARN] <<< Exiting TestChooseData(t=%v)", t)
}

func TestFilterFlags(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFilterFlags(t=%v)", t)
	result := filterFlags("text/html ")
	assert.Equal(t, "text/html", result)

	result = filterFlags("text/html;")
	assert.Equal(t, "text/html", result)
	logritelog.Default().Printf("[WARN] <<< Exiting TestFilterFlags(t=%v)", t)
}

func TestFunctionName(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFunctionName(t=%v)", t)
	assert.Regexp(t, `^(.*/vendor/)?github.com/gin-gonic/gin.somefunction$`, nameOfFunction(somefunction))
	logritelog.Default().Printf("[WARN] <<< Exiting TestFunctionName(t=%v)", t)
}

func somefunction() {
	logritelog.Default().Printf("[ERROR] >>> Entering somefunction()")
	// this empty function is used by TestFunctionName()
	logritelog.Default().Printf("[WARN] <<< Exiting somefunction()")
}

func TestJoinPaths(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestJoinPaths(t=%v)", t)
	assert.Empty(t, joinPaths("", ""))
	assert.Equal(t, "/", joinPaths("", "/"))
	assert.Equal(t, "/a", joinPaths("/a", ""))
	assert.Equal(t, "/a/", joinPaths("/a/", ""))
	assert.Equal(t, "/a/", joinPaths("/a/", "/"))
	assert.Equal(t, "/a/", joinPaths("/a", "/"))
	assert.Equal(t, "/a/hola", joinPaths("/a", "/hola"))
	assert.Equal(t, "/a/hola", joinPaths("/a/", "/hola"))
	assert.Equal(t, "/a/hola/", joinPaths("/a/", "/hola/"))
	assert.Equal(t, "/a/hola/", joinPaths("/a/", "/hola//"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestJoinPaths(t=%v)", t)
}

type bindTestStruct struct {
	Foo string `form:"foo" binding:"required"`
	Bar int    `form:"bar" binding:"min=4"`
}

func TestBindMiddleware(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestBindMiddleware(t=%v)", t)
	var value *bindTestStruct
	var called bool
	router := New()
	router.GET("/", Bind(bindTestStruct{}), func(c *Context) {
		called = true
		logritelog.Default().Printf("[DEBUG] TestBindMiddleware(t=%v): called → %v", t, called)
		value = c.MustGet(BindKey).(*bindTestStruct)
	})
	PerformRequest(router, http.MethodGet, "/?foo=hola&bar=10")
	assert.True(t, called)
	assert.Equal(t, "hola", value.Foo)
	assert.Equal(t, 10, value.Bar)

		logritelog.Default().Printf("[DEBUG] TestBindMiddleware(t=%v): value → %v", t, value)
	called = false
	PerformRequest(router, http.MethodGet, "/?foo=hola&bar=1")
	assert.False(t, called)

	assert.Panics(t, func() {
		Bind(&bindTestStruct{})
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestBindMiddleware(t=%v)", t)
}

func TestMarshalXMLforH(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMarshalXMLforH(t=%v)", t)
	h := H{
		"": "test",
	}
	var b bytes.Buffer
	enc := xml.NewEncoder(&b)
	var x xml.StartElement
	e := h.MarshalXML(enc, x)
	assert.Error(t, e)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMarshalXMLforH(t=%v)", t)
}

func TestMarshalXMLforHSuccess(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMarshalXMLforHSuccess(t=%v)", t)
	h := H{
		"key1": "value1",
		"key2": 123,
	}
	data, err := xml.Marshal(h)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<key1>value1</key1>")
	assert.Contains(t, string(data), "<key2>123</key2>")
	logritelog.Default().Printf("[WARN] <<< Exiting TestMarshalXMLforHSuccess(t=%v)", t)
}

func TestIsASCII(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestIsASCII(t=%v)", t)
	assert.True(t, isASCII("test"))
	assert.False(t, isASCII("🧡💛💚💙💜"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestIsASCII(t=%v)", t)
}

func TestSafeInt8(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestSafeInt8(t=%v)", t)
	assert.Equal(t, int8(100), safeInt8(100))
	assert.Equal(t, int8(math.MaxInt8), safeInt8(int(math.MaxInt8)+123))
	logritelog.Default().Printf("[WARN] <<< Exiting TestSafeInt8(t=%v)", t)
}

func TestSafeUint16(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestSafeUint16(t=%v)", t)
	assert.Equal(t, uint16(100), safeUint16(100))
	assert.Equal(t, uint16(math.MaxUint16), safeUint16(int(math.MaxUint16)+123))
	logritelog.Default().Printf("[WARN] <<< Exiting TestSafeUint16(t=%v)", t)
}
