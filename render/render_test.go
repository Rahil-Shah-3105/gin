// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import (
	logritelog "log"
	logritetime "time"
)

import (
	"encoding/xml"
	"errors"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	testdata "github.com/gin-gonic/gin/testdata/protoexample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
)

func TestRenderJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJSON(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}

	(JSON{data}).WriteContentType(w)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	err := (JSON{data}).Render(w)

	require.NoError(t, err)
	assert.JSONEq(t, "{\"foo\":\"bar\",\"html\":\"\\u003cb\\u003e\"}", w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJSON(t=%v)", t)
}

func TestRenderJSONError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJSONError(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	require.Error(t, (JSON{data}).Render(w))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJSONError(t=%v)", t)
}

func TestRenderIndentedJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderIndentedJSON(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
		"bar": "foo",
	}

	err := (IndentedJSON{data}).Render(w)

	require.NoError(t, err)
	assert.JSONEq(t, "{\n    \"bar\": \"foo\",\n    \"foo\": \"bar\"\n}", w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderIndentedJSON(t=%v)", t)
}

func TestRenderIndentedJSONPanics(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderIndentedJSONPanics(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (IndentedJSON{data}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderIndentedJSONPanics(t=%v)", t)
}

func TestRenderSecureJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderSecureJSON(t=%v)", t)
	w1 := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(SecureJSON{"while(1);", data}).WriteContentType(w1)
	assert.Equal(t, "application/json; charset=utf-8", w1.Header().Get("Content-Type"))

	err1 := (SecureJSON{"while(1);", data}).Render(w1)

	require.NoError(t, err1)
	assert.JSONEq(t, "{\"foo\":\"bar\"}", w1.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w1.Header().Get("Content-Type"))

	w2 := httptest.NewRecorder()
	datas := []map[string]any{{
		"foo": "bar",
	}, {
		"bar": "foo",
	}}

	err2 := (SecureJSON{"while(1);", datas}).Render(w2)
	require.NoError(t, err2)
	assert.Equal(t, "while(1);[{\"foo\":\"bar\"},{\"bar\":\"foo\"}]", w2.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w2.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderSecureJSON(t=%v)", t)
}

func TestRenderSecureJSONFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderSecureJSONFail(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (SecureJSON{"while(1);", data}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderSecureJSONFail(t=%v)", t)
}

func TestRenderJsonpJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJsonpJSON(t=%v)", t)
	w1 := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}

	(JsonpJSON{"x", data}).WriteContentType(w1)
	assert.Equal(t, "application/javascript; charset=utf-8", w1.Header().Get("Content-Type"))

	err1 := (JsonpJSON{"x", data}).Render(w1)

	require.NoError(t, err1)
	assert.Equal(t, "x({\"foo\":\"bar\"});", w1.Body.String())
	assert.Equal(t, "application/javascript; charset=utf-8", w1.Header().Get("Content-Type"))

	w2 := httptest.NewRecorder()
	datas := []map[string]any{{
		"foo": "bar",
	}, {
		"bar": "foo",
	}}

	err2 := (JsonpJSON{"x", datas}).Render(w2)
	require.NoError(t, err2)
	assert.Equal(t, "x([{\"foo\":\"bar\"},{\"bar\":\"foo\"}]);", w2.Body.String())
	assert.Equal(t, "application/javascript; charset=utf-8", w2.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJsonpJSON(t=%v)", t)
}

type errorWriter struct {
	bufString    string
	ErrThreshold int // 1-based threshold. If 1, errors on the 1st Write call.
	writeCount   int
	*httptest.ResponseRecorder
}

var _ http.ResponseWriter = (*errorWriter)(nil)

func (w *errorWriter) Header() http.Header {
	logritelog.Default().Printf("[ERROR] >>> Entering Header()")
	if w.ResponseRecorder == nil {
		w.ResponseRecorder = httptest.NewRecorder()
	}
	logritelog.Default().Printf("[WARN] <<< Exiting Header()")
	return w.ResponseRecorder.Header()
}

func (w *errorWriter) WriteHeader(statusCode int) {
	logritelog.Default().Printf("[ERROR] >>> Entering WriteHeader(statusCode=%v)", statusCode)
	if w.ResponseRecorder == nil {
		w.ResponseRecorder = httptest.NewRecorder()
	}
	w.ResponseRecorder.WriteHeader(statusCode)
	logritelog.Default().Printf("[WARN] <<< Exiting WriteHeader(statusCode=%v)", statusCode)
}

func (w *errorWriter) Write(buf []byte) (int, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Write(buf=%v)", buf)
	w.writeCount++
	if (w.bufString != "" && string(buf) == w.bufString) || (w.ErrThreshold > 0 && w.writeCount >= w.ErrThreshold) {
		return 0, errors.New(`write error`)
	}
	if w.ResponseRecorder == nil {
		w.ResponseRecorder = httptest.NewRecorder()
	}
	logritelog.Default().Printf("[WARN] <<< Exiting Write(buf=%v)", buf)
	return w.ResponseRecorder.Write(buf)
}

func (w *errorWriter) reset() {
	logritelog.Default().Printf("[ERROR] >>> Entering reset()")
	w.writeCount = 0
	w.ResponseRecorder = httptest.NewRecorder()
	logritelog.Default().Printf("[DEBUG] reset(): w.writeCount → %v", w.writeCount)
logritelog.Default().Printf("[WARN] <<< Exiting reset()")
}

func TestRenderJsonpJSONError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJsonpJSONError(t=%v)", t)
	ew := &errorWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}

	jsonpJSON := JsonpJSON{
		Callback: "foo",
		Data: map[string]string{
			"foo": "bar",
		},
	}

	// error was returned while writing callback
	ew.reset()
	ew.ErrThreshold = 1
	logritelog.Default().Printf("[DEBUG] TestRenderJsonpJSONError(t=%v): ew.ErrThreshold → %v", t, ew.ErrThreshold)
	err := jsonpJSON.Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())

	// error was returned while writing "("
	ew.reset()
	ew.ErrThreshold = 2
	err = jsonpJSON.Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())

	logritelog.Default().Printf("[DEBUG] TestRenderJsonpJSONError(t=%v): err → %v", t, err)
	// error was returned while writing data
	ew.reset()
	ew.ErrThreshold = 3
	err = jsonpJSON.Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())

	// error was returned while writing ");"
	ew.reset()
	ew.ErrThreshold = 4
	err = jsonpJSON.Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJsonpJSONError(t=%v)", t)
}

func TestRenderJsonpJSONError2(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJsonpJSONError2(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo": "bar",
	}
	(JsonpJSON{"", data}).WriteContentType(w)
	assert.Equal(t, "application/javascript; charset=utf-8", w.Header().Get("Content-Type"))

	e := (JsonpJSON{"", data}).Render(w)
	require.NoError(t, e)

	assert.JSONEq(t, "{\"foo\":\"bar\"}", w.Body.String())
	assert.Equal(t, "application/javascript; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJsonpJSONError2(t=%v)", t)
}

func TestRenderJsonpJSONFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderJsonpJSONFail(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	err := (JsonpJSON{"x", data}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderJsonpJSONFail(t=%v)", t)
}

func TestRenderAsciiJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderAsciiJSON(t=%v)", t)
	w1 := httptest.NewRecorder()
	data1 := map[string]any{
		"lang": "GO语言",
		"tag":  "<br>",
	}

	err := (AsciiJSON{data1}).Render(w1)

	require.NoError(t, err)
	assert.JSONEq(t, "{\"lang\":\"GO\\u8bed\\u8a00\",\"tag\":\"\\u003cbr\\u003e\"}", w1.Body.String())
	assert.Equal(t, "application/json", w1.Header().Get("Content-Type"))

	w2 := httptest.NewRecorder()
	data2 := 3.1415926

	err = (AsciiJSON{data2}).Render(w2)
	require.NoError(t, err)
	assert.Equal(t, "3.1415926", w2.Body.String())
	logritelog.Default().Printf("[DEBUG] TestRenderAsciiJSON(t=%v): err → %v", t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderAsciiJSON(t=%v)", t)
}

func TestRenderAsciiJSONFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderAsciiJSONFail(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	// json: unsupported type: chan int
	require.Error(t, (AsciiJSON{data}).Render(w))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderAsciiJSONFail(t=%v)", t)
}

func TestRenderPureJSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderPureJSON(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}
	err := (PureJSON{data}).Render(w)
	require.NoError(t, err)
	assert.JSONEq(t, "{\"foo\":\"bar\",\"html\":\"<b>\"}\n", w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderPureJSON(t=%v)", t)
}

type xmlmap map[string]any

// Allows type H to be used with xml.Marshal
func (h xmlmap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalXML(e=%v,start=%v)", e, start)
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalXML(e=%v,start=%v)", e, start)
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	logritelog.Default().Printf("[DEBUG] MarshalXML(e=%v,start=%v): start.Name → %v", e, start, start.Name)
	if err := e.EncodeToken(start); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in MarshalXML(e=%v,start=%v): %v", e, start, err)
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	logritelog.Default().Printf("[WARN] <<< Exiting MarshalXML(e=%v,start=%v)", e, start)
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestRenderYAML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderYAML(t=%v)", t)
	w := httptest.NewRecorder()
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderYAML(t=%v)", t)
	data := `
a : Easy!
b:
	c: 2
	d: [3, 4]
	`
	(YAML{data}).WriteContentType(w)
	assert.Equal(t, "application/yaml; charset=utf-8", w.Header().Get("Content-Type"))

	err := (YAML{data}).Render(w)
	require.NoError(t, err)

	// With github.com/goccy/go-yaml, the output format is different from gopkg.in/yaml.v3
	// We're checking that the output contains the expected data, not the exact formatting
	output := w.Body.String()
	assert.Contains(t, output, "a : Easy!")
	assert.Contains(t, output, "b:")
	assert.Contains(t, output, "c: 2")
	assert.Contains(t, output, "d: [3, 4]")
	assert.Equal(t, "application/yaml; charset=utf-8", w.Header().Get("Content-Type"))
}

type fail struct{}

// Hook MarshalYAML
func (ft *fail) MarshalYAML() (any, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering MarshalYAML()")
	logritelog.Default().Printf("[WARN] <<< Exiting MarshalYAML()")
	return nil, errors.New("fail")
}

func TestRenderYAMLFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderYAMLFail(t=%v)", t)
	w := httptest.NewRecorder()
	err := (YAML{&fail{}}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderYAMLFail(t=%v)", t)
}

func TestRenderTOML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderTOML(t=%v)", t)
	w := httptest.NewRecorder()
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}
	(TOML{data}).WriteContentType(w)
	assert.Equal(t, "application/toml; charset=utf-8", w.Header().Get("Content-Type"))

	err := (TOML{data}).Render(w)
	require.NoError(t, err)
	assert.Equal(t, "foo = 'bar'\nhtml = '<b>'\n", w.Body.String())
	assert.Equal(t, "application/toml; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderTOML(t=%v)", t)
}

func TestRenderTOMLFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderTOMLFail(t=%v)", t)
	w := httptest.NewRecorder()
	err := (TOML{net.IPv4bcast}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderTOMLFail(t=%v)", t)
}

// test Protobuf rendering
func TestRenderProtoBuf(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderProtoBuf(t=%v)", t)
	w := httptest.NewRecorder()
	reps := []int64{int64(1), int64(2)}
	label := "test"
	data := &testdata.Test{
		Label: &label,
		Reps:  reps,
	}

	(ProtoBuf{data}).WriteContentType(w)
	protoData, err := proto.Marshal(data)
	require.NoError(t, err)
	assert.Equal(t, "application/x-protobuf", w.Header().Get("Content-Type"))

	err = (ProtoBuf{data}).Render(w)

	logritelog.Default().Printf("[DEBUG] TestRenderProtoBuf(t=%v): err → %v", t, err)
	require.NoError(t, err)
	assert.Equal(t, string(protoData), w.Body.String())
	assert.Equal(t, "application/x-protobuf", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderProtoBuf(t=%v)", t)
}

func TestRenderProtoBufFail(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderProtoBufFail(t=%v)", t)
	w := httptest.NewRecorder()
	data := &testdata.Test{}
	err := (ProtoBuf{data}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderProtoBufFail(t=%v)", t)
}

func TestRenderBSON(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderBSON(t=%v)", t)
	w := httptest.NewRecorder()
	reps := []int64{int64(1), int64(2)}
	type mystruct struct {
		Label string
		Reps  []int64
	}

	data := &mystruct{
		Label: "test",
		Reps:  reps,
	}

	(BSON{data}).WriteContentType(w)
	bsonData, err := bson.Marshal(data)
	require.NoError(t, err)
	assert.Equal(t, "application/bson", w.Header().Get("Content-Type"))

	err = (BSON{data}).Render(w)

	logritelog.Default().Printf("[DEBUG] TestRenderBSON(t=%v): err → %v", t, err)
	require.NoError(t, err)
	assert.Equal(t, bsonData, w.Body.Bytes())
	assert.Equal(t, "application/bson", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderBSON(t=%v)", t)
}

func TestRenderBSONError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderBSONError(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	err := (BSON{data}).Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderBSONError(t=%v)", t)
}

func TestRenderBSONWriteError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderBSONWriteError(t=%v)", t)
	type testStruct struct {
		Value string
	}
	data := &testStruct{Value: "test"}

	ew := &errorWriter{
		ErrThreshold:     1,
		ResponseRecorder: httptest.NewRecorder(),
	}

	err := (BSON{data}).Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderBSONWriteError(t=%v)", t)
}

func TestRenderXML(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderXML(t=%v)", t)
	w := httptest.NewRecorder()
	data := xmlmap{
		"foo": "bar",
	}

	(XML{data}).WriteContentType(w)
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))

	err := (XML{data}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, "<map><foo>bar</foo></map>", w.Body.String())
	assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderXML(t=%v)", t)
}

func TestRenderXMLError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderXMLError(t=%v)", t)
	w := httptest.NewRecorder()
	data := make(chan int)

	err := (XML{data}).Render(w)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xml: unsupported type: chan int")
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderXMLError(t=%v)", t)
}

func TestRenderPDF(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderPDF(t=%v)", t)
	w := httptest.NewRecorder()
	data := []byte("%Test pdf content")

	pdf := PDF{data}

	pdf.WriteContentType(w)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))

	err := pdf.Render(w)
	require.NoError(t, err)

	assert.Equal(t, data, w.Body.Bytes())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderPDF(t=%v)", t)
}

func TestRenderRedirect(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderRedirect(t=%v)", t)
	__logStart := logritetime.Now()
	req, err := http.NewRequest(http.MethodGet, "/test-redirect", nil)
	require.NoError(t, err)

	data1 := Redirect{
		Code:     http.StatusMovedPermanently,
		Request:  req,
		Location: "/new/location",
	}

	w := httptest.NewRecorder()
	err = data1.Render(w)
	require.NoError(t, err)

	logritelog.Default().Printf("[DEBUG] TestRenderRedirect(t=%v): err → %v", t, err)
	data2 := Redirect{
		Code:     http.StatusOK,
		Request:  req,
		Location: "/new/location",
	}

	w = httptest.NewRecorder()
	assert.PanicsWithValue(t, "Cannot redirect with status code 200", func() {
		err := data2.Render(w)
		require.NoError(t, err)
	})

	data3 := Redirect{
		Code:     http.StatusCreated,
		Request:  req,
		Location: "/new/location",
	}

	w = httptest.NewRecorder()
	err = data3.Render(w)
	require.NoError(t, err)

	// only improve coverage
	data2.WriteContentType(w)
	logritelog.Default().Printf("[ERROR] TestRenderRedirect(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestRenderRedirect(t=%v)", t)
}

func TestRenderData(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderData(t=%v)", t)
	w := httptest.NewRecorder()
	data := []byte("#!PNG some raw data")

	err := (Data{
		ContentType: "image/png",
		Data:        data,
	}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, "#!PNG some raw data", w.Body.String())
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Equal(t, "19", w.Header().Get("Content-Length"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderData(t=%v)", t)
}

func TestRenderDataContentLength(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderDataContentLength(t=%v)", t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, err := strconv.Atoi(r.URL.Query().Get("size"))
		assert.NoError(t, err)

		data := Data{
			ContentType: "application/octet-stream",
			Data:        make([]byte, size),
		}
		assert.NoError(t, data.Render(w))
	}))
	t.Cleanup(srv.Close)

	for _, size := range []int{0, 1, 100, 100_000} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			resp, err := http.Get(srv.URL + "?size=" + strconv.Itoa(size))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
			assert.Equal(t, strconv.Itoa(size), resp.Header.Get("Content-Length"))

			actual, err := io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)
			assert.EqualValues(t, size, actual)
		})
	}
	logritelog.Default().Printf("[ERROR] TestRenderDataContentLength(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestRenderDataContentLength(t=%v)", t)
}

func TestRenderDataError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderDataError(t=%v)", t)
	ew := &errorWriter{
		ErrThreshold:     1,
		ResponseRecorder: httptest.NewRecorder(),
	}
	data := []byte("#!PNG some raw data")

	err := (Data{
		ContentType: "image/png",
		Data:        data,
	}).Render(ew)

	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderDataError(t=%v)", t)
}

func TestRenderString(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderString(t=%v)", t)
	w := httptest.NewRecorder()

	(String{
		Format: "hello %s %d",
		Data:   []any{},
	}).WriteContentType(w)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	err := (String{
		Format: "hola %s %d",
		Data:   []any{"manu", 2},
	}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, "hola manu 2", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderString(t=%v)", t)
}

func TestRenderStringLenZero(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderStringLenZero(t=%v)", t)
	w := httptest.NewRecorder()

	err := (String{
		Format: "hola %s %d",
		Data:   []any{},
	}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, "hola %s %d", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderStringLenZero(t=%v)", t)
}

func TestRenderHTMLTemplate(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLTemplate(t=%v)", t)
	w := httptest.NewRecorder()
	templ := template.Must(template.New("t").Parse(`Hello {{.name}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("t", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)

	require.NoError(t, err)
	assert.Equal(t, "Hello alexandernyquist", w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLTemplate(t=%v)", t)
}

func TestRenderHTMLTemplateEmptyName(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLTemplateEmptyName(t=%v)", t)
	w := httptest.NewRecorder()
	templ := template.Must(template.New("").Parse(`Hello {{.name}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)

	require.NoError(t, err)
	assert.Equal(t, "Hello alexandernyquist", w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLTemplateEmptyName(t=%v)", t)
}

func TestRenderHTMLDebugFiles(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLDebugFiles(t=%v)", t)
	w := httptest.NewRecorder()
	htmlRender := HTMLDebug{
		Files:      []string{"../testdata/template/hello.tmpl"},
		Glob:       "",
		FileSystem: nil,
		Patterns:   nil,
		Delims:     Delims{Left: "{[{", Right: "}]}"},
		FuncMap:    nil,
	}
	instance := htmlRender.Instance("hello.tmpl", map[string]any{
		"name": "thinkerou",
	})

	err := instance.Render(w)

	require.NoError(t, err)
	assert.Equal(t, "<h1>Hello thinkerou</h1>", w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLDebugFiles(t=%v)", t)
}

func TestRenderHTMLDebugGlob(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLDebugGlob(t=%v)", t)
	w := httptest.NewRecorder()
	htmlRender := HTMLDebug{
		Files:      nil,
		Glob:       "../testdata/template/hello*",
		FileSystem: nil,
		Patterns:   nil,
		Delims:     Delims{Left: "{[{", Right: "}]}"},
		FuncMap:    nil,
	}
	instance := htmlRender.Instance("hello.tmpl", map[string]any{
		"name": "thinkerou",
	})

	err := instance.Render(w)

	require.NoError(t, err)
	assert.Equal(t, "<h1>Hello thinkerou</h1>", w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLDebugGlob(t=%v)", t)
}

func TestRenderHTMLDebugFS(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLDebugFS(t=%v)", t)
	w := httptest.NewRecorder()
	htmlRender := HTMLDebug{
		Files:      nil,
		Glob:       "",
		FileSystem: http.Dir("../testdata/template"),
		Patterns:   []string{"hello.tmpl"},
		Delims:     Delims{Left: "{[{", Right: "}]}"},
		FuncMap:    nil,
	}
	instance := htmlRender.Instance("hello.tmpl", map[string]any{
		"name": "thinkerou",
	})

	err := instance.Render(w)

	require.NoError(t, err)
	assert.Equal(t, "<h1>Hello thinkerou</h1>", w.Body.String())
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLDebugFS(t=%v)", t)
}

func TestRenderHTMLDebugPanics(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLDebugPanics(t=%v)", t)
	htmlRender := HTMLDebug{
		Files:      nil,
		Glob:       "",
		FileSystem: nil,
		Patterns:   nil,
		Delims:     Delims{"{{", "}}"},
		FuncMap:    nil,
	}
	assert.Panics(t, func() { htmlRender.Instance("", nil) })
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLDebugPanics(t=%v)", t)
}

func TestRenderHTMLTemplateError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLTemplateError(t=%v)", t)
	w := httptest.NewRecorder()
	templ := template.Must(template.New("t").Parse(`Hello {{if .name}}{{.name.DoesNotExist}}{{end}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("t", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLTemplateError(t=%v)", t)
}

func TestRenderHTMLTemplateExecuteError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderHTMLTemplateExecuteError(t=%v)", t)
	w := httptest.NewRecorder()
	templ := template.Must(template.New("t").Parse(`Hello {{.name.invalid}}`))

	htmlRender := HTMLProduction{Template: templ}
	instance := htmlRender.Instance("t", map[string]any{
		"name": "alexandernyquist",
	})

	err := instance.Render(w)
	require.Error(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderHTMLTemplateExecuteError(t=%v)", t)
}

func TestRenderReader(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderReader(t=%v)", t)
	w := httptest.NewRecorder()

	body := "#!PNG some raw data"
	headers := make(map[string]string)
	headers["Content-Disposition"] = `attachment; filename="filename.png"`
	headers["x-request-id"] = "requestId"

	err := (Reader{
		ContentLength: int64(len(body)),
		ContentType:   "image/png",
		Reader:        strings.NewReader(body),
		Headers:       headers,
	}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, body, w.Body.String())
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Equal(t, strconv.Itoa(len(body)), w.Header().Get("Content-Length"))
	assert.Equal(t, headers["Content-Disposition"], w.Header().Get("Content-Disposition"))
	assert.Equal(t, headers["x-request-id"], w.Header().Get("x-request-id"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderReader(t=%v)", t)
}

func TestRenderReaderNoContentLength(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderReaderNoContentLength(t=%v)", t)
	w := httptest.NewRecorder()

	body := "#!PNG some raw data"
	headers := make(map[string]string)
	headers["Content-Disposition"] = `attachment; filename="filename.png"`
	headers["x-request-id"] = "requestId"

	err := (Reader{
		ContentLength: -1,
		ContentType:   "image/png",
		Reader:        strings.NewReader(body),
		Headers:       headers,
	}).Render(w)

	require.NoError(t, err)
	assert.Equal(t, body, w.Body.String())
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotContains(t, "Content-Length", w.Header())
	assert.Equal(t, headers["Content-Disposition"], w.Header().Get("Content-Disposition"))
	assert.Equal(t, headers["x-request-id"], w.Header().Get("x-request-id"))
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderReaderNoContentLength(t=%v)", t)
}

func TestRenderWriteError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRenderWriteError(t=%v)", t)
	data := []any{"value1", "value2"}
	prefix := "my-prefix:"
	r := SecureJSON{Data: data, Prefix: prefix}
	ew := &errorWriter{
		ErrThreshold:     1,
		ResponseRecorder: httptest.NewRecorder(),
	}
	err := r.Render(ew)
	require.Error(t, err)
	assert.Equal(t, "write error", err.Error())
	logritelog.Default().Printf("[WARN] <<< Exiting TestRenderWriteError(t=%v)", t)
}
