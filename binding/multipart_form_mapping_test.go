// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import (
	logritelog "log"
	logritetime "time"
)

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormMultipartBindingBindOneFile(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormMultipartBindingBindOneFile(t=%v)", t)
	var s struct {
		FileValue   multipart.FileHeader     `form:"file"`
		FilePtr     *multipart.FileHeader    `form:"file"`
		SliceValues []multipart.FileHeader   `form:"file"`
		SlicePtrs   []*multipart.FileHeader  `form:"file"`
		ArrayValues [1]multipart.FileHeader  `form:"file"`
		ArrayPtrs   [1]*multipart.FileHeader `form:"file"`
	}
	file := testFile{"file", "file1", []byte("hello")}

	req := createRequestMultipartFiles(t, file)
	err := FormMultipart.Bind(req, &s)
	require.NoError(t, err)

	assertMultipartFileHeader(t, &s.FileValue, file)
	assertMultipartFileHeader(t, s.FilePtr, file)
	assert.Len(t, s.SliceValues, 1)
	assertMultipartFileHeader(t, &s.SliceValues[0], file)
	assert.Len(t, s.SlicePtrs, 1)
	assertMultipartFileHeader(t, s.SlicePtrs[0], file)
	assertMultipartFileHeader(t, &s.ArrayValues[0], file)
	assertMultipartFileHeader(t, s.ArrayPtrs[0], file)
	logritelog.Default().Printf("[WARN] <<< Exiting TestFormMultipartBindingBindOneFile(t=%v)", t)
}

func TestFormMultipartBindingBindTwoFiles(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormMultipartBindingBindTwoFiles(t=%v)", t)
	var s struct {
		SliceValues []multipart.FileHeader   `form:"file"`
		SlicePtrs   []*multipart.FileHeader  `form:"file"`
		ArrayValues [2]multipart.FileHeader  `form:"file"`
		ArrayPtrs   [2]*multipart.FileHeader `form:"file"`
	}
	files := []testFile{
		{"file", "file1", []byte("hello")},
		{"file", "file2", []byte("world")},
	}

	req := createRequestMultipartFiles(t, files...)
	err := FormMultipart.Bind(req, &s)
	require.NoError(t, err)

	assert.Len(t, s.SliceValues, len(files))
	assert.Len(t, s.SlicePtrs, len(files))
	assert.Len(t, s.ArrayValues, len(files))
	assert.Len(t, s.ArrayPtrs, len(files))

	for i, file := range files {
		assertMultipartFileHeader(t, &s.SliceValues[i], file)
		assertMultipartFileHeader(t, s.SlicePtrs[i], file)
		assertMultipartFileHeader(t, &s.ArrayValues[i], file)
		assertMultipartFileHeader(t, s.ArrayPtrs[i], file)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestFormMultipartBindingBindTwoFiles(t=%v)", t)
}

func TestFormMultipartBindingBindError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestFormMultipartBindingBindError(t=%v)", t)
	files := []testFile{
		{"file", "file1", []byte("hello")},
		{"file", "file2", []byte("world")},
	}

	for _, tt := range []struct {
		name string
		s    any
	}{
		{"wrong type", &struct {
			Files int `form:"file"`
		}{}},
		{"wrong array size", &struct {
			Files [1]*multipart.FileHeader `form:"file"`
		}{}},
		{"wrong slice type", &struct {
			Files []int `form:"file"`
		}{}},
	} {
		req := createRequestMultipartFiles(t, files...)
		err := FormMultipart.Bind(req, tt.s)
		require.Error(t, err)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestFormMultipartBindingBindError(t=%v)", t)
}

type testFile struct {
	Fieldname string
	Filename  string
	Content   []byte
}

func createRequestMultipartFiles(t *testing.T, files ...testFile) *http.Request {
	logritelog.Default().Printf("[ERROR] >>> Entering createRequestMultipartFiles(t=%v,files=%v)", t, files)
	__logStart := logritetime.Now()
	var body bytes.Buffer

	mw := multipart.NewWriter(&body)
	for _, file := range files {
		fw, err := mw.CreateFormFile(file.Fieldname, file.Filename)
		require.NoError(t, err)

		n, err := fw.Write(file.Content)
		require.NoError(t, err)
		assert.Equal(t, len(file.Content), n)
	}
	err := mw.Close()
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/", &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", MIMEMultipartPOSTForm+"; boundary="+mw.Boundary())
	logritelog.Default().Printf("[ERROR] createRequestMultipartFiles(t,files)=%v,%v: external service call took %v ms", t, files, logritetime.Since(__logStart).Milliseconds())
	logritelog.Default().Printf("[WARN] <<< Exiting createRequestMultipartFiles(t=%v,files=%v)", t, files)
	return req
}

func assertMultipartFileHeader(t *testing.T, fh *multipart.FileHeader, file testFile) {
	logritelog.Default().Printf("[ERROR] >>> Entering assertMultipartFileHeader(t=%v,fh=%v,file=%v)", t, fh, file)
	assert.Equal(t, file.Filename, fh.Filename)
	assert.Equal(t, int64(len(file.Content)), fh.Size)

	fl, err := fh.Open()
	require.NoError(t, err)

	body, err := io.ReadAll(fl)
	require.NoError(t, err)
	assert.Equal(t, string(file.Content), string(body))

	err = fl.Close()
	require.NoError(t, err)
	logritelog.Default().Printf("[DEBUG] assertMultipartFileHeader(t=%v,fh=%v,file=%v): err → %v", t, fh, file, err)
	logritelog.Default().Printf("[WARN] <<< Exiting assertMultipartFileHeader(t=%v,fh=%v,file=%v)", t, fh, file)
}
