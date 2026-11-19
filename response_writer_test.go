// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import (
	logritelog "log"
	logritetime "time"
)

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponseWriterFlushWithFlusher verifies Flush() calls the underlying Flusher.
func TestResponseWriterFlushWithFlusher(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterFlushWithFlusher(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{ResponseWriter: testWriter}
	writer.Flush()
	assert.True(t, testWriter.Flushed)
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterFlushWithFlusher(t=%v)", t)
}

// TestResponseWriterFlushWithNonFlusher verifies Flush() is a no-op
// when the underlying ResponseWriter does not implement http.Flusher.
// Guards against the panic reported in https://github.com/gin-gonic/gin/issues/4460
func TestResponseWriterFlushWithNonFlusher(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterFlushWithNonFlusher(t=%v)", t)
	nonFlusher := &nonFlusherWriter{header: http.Header{}}
	writer := &responseWriter{ResponseWriter: nonFlusher}
	require.NotPanics(t, func() {
		writer.Flush()
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterFlushWithNonFlusher(t=%v)", t)
}

// nonFlusherWriter is a minimal http.ResponseWriter that does NOT implement http.Flusher.
type nonFlusherWriter struct {
	header http.Header
}

func (w *nonFlusherWriter) Header() http.Header         {
    logritelog.Default().Printf("[ERROR] >>> Entering Header()")
    logritelog.Default().Printf("[WARN] <<< Exiting Header()")
    return w.header
}
func (w *nonFlusherWriter) Write(b []byte) (int, error) {
    logritelog.Default().Printf("[ERROR] >>> Entering Write(b=%v)", b)
    logritelog.Default().Printf("[WARN] <<< Exiting Write(b=%v)", b)
    return len(b), nil
}
func (w *nonFlusherWriter) WriteHeader(code int)        {}

var (
	_ ResponseWriter      = &responseWriter{}
	_ http.ResponseWriter = &responseWriter{}
	_ http.ResponseWriter = ResponseWriter(&responseWriter{})
	_ http.Hijacker       = ResponseWriter(&responseWriter{})
	_ http.Flusher        = ResponseWriter(&responseWriter{})
	_ http.CloseNotifier  = ResponseWriter(&responseWriter{})
)

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	SetMode(TestMode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

func TestResponseWriterUnwrap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterUnwrap(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{ResponseWriter: testWriter}
	assert.Same(t, testWriter, writer.Unwrap())
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterUnwrap(t=%v)", t)
}

func TestResponseWriterReset(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterReset(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	var w ResponseWriter = writer

	writer.reset(testWriter)
	assert.Equal(t, -1, writer.size)
	assert.Equal(t, http.StatusOK, writer.status)
	assert.Equal(t, testWriter, writer.ResponseWriter)
	assert.Equal(t, -1, w.Size())
	assert.Equal(t, http.StatusOK, w.Status())
	assert.False(t, w.Written())
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterReset(t=%v)", t)
}

func TestResponseWriterWriteHeader(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterWriteHeader(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(testWriter)
	w := ResponseWriter(writer)

	w.WriteHeader(http.StatusMultipleChoices)
	assert.False(t, w.Written())
	assert.Equal(t, http.StatusMultipleChoices, w.Status())
	assert.NotEqual(t, http.StatusMultipleChoices, testWriter.Code)

	w.WriteHeader(-1)
	assert.Equal(t, http.StatusMultipleChoices, w.Status())
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterWriteHeader(t=%v)", t)
}

func TestResponseWriterWriteHeadersNow(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterWriteHeadersNow(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(testWriter)
	w := ResponseWriter(writer)

	w.WriteHeader(http.StatusMultipleChoices)
	w.WriteHeaderNow()

	assert.True(t, w.Written())
	assert.Equal(t, 0, w.Size())
	assert.Equal(t, http.StatusMultipleChoices, testWriter.Code)

	writer.size = 10
	logritelog.Default().Printf("[DEBUG] TestResponseWriterWriteHeadersNow(t=%v): writer.size → %v", t, writer.size)
	w.WriteHeaderNow()
	assert.Equal(t, 10, w.Size())
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterWriteHeadersNow(t=%v)", t)
}

func TestResponseWriterWrite(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterWrite(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(testWriter)
	w := ResponseWriter(writer)

	n, err := w.Write([]byte("hola"))
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, w.Size())
	assert.Equal(t, http.StatusOK, w.Status())
	assert.Equal(t, http.StatusOK, testWriter.Code)
	assert.Equal(t, "hola", testWriter.Body.String())
	require.NoError(t, err)

	n, err = w.Write([]byte(" adios"))
	assert.Equal(t, 6, n)
	assert.Equal(t, 10, w.Size())
	assert.Equal(t, "hola adios", testWriter.Body.String())
	require.NoError(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterWrite(t=%v)", t)
}

func TestResponseWriterHijack(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterHijack(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(testWriter)
	w := ResponseWriter(writer)

	// httptest.ResponseRecorder doesn't implement http.Hijacker; return
	// http.ErrNotSupported instead of panicking (#4638). On unsupported the
	// writer state stays untouched so the handler can still emit a normal
	// HTTP response as a fallback.
	conn, buf, err := w.Hijack()
	assert.Nil(t, conn)
	assert.Nil(t, buf)
	require.ErrorIs(t, err, http.ErrNotSupported)
	assert.False(t, w.Written())

	// CloseNotify on a non-CloseNotifier returns nil instead of panicking.
	assert.Nil(t, w.CloseNotify())

	w.Flush()
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterHijack(t=%v)", t)
}

type mockHijacker struct {
	*httptest.ResponseRecorder
	hijacked bool
}

// Hijack implements the http.Hijacker interface. It just records that it was called.
func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Hijack()")
	m.hijacked = true
	logritelog.Default().Printf("[DEBUG] Hijack(): m.hijacked → %v", m.hijacked)
	logritelog.Default().Printf("[WARN] <<< Exiting Hijack()")
	return nil, nil, nil
}

func TestResponseWriterHijackAfterWrite(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterHijackAfterWrite(t=%v)", t)
	tests := []struct {
		name                      string
		action                    func(w ResponseWriter) error // Action to perform before hijacking
		expectWrittenBeforeHijack bool
		expectHijackSuccess       bool
		expectWrittenAfterHijack  bool
		expectError               error
	}{
		{
			name:                      "hijack before write should succeed",
			action:                    func(w ResponseWriter) error { return nil },
			expectWrittenBeforeHijack: false,
			expectHijackSuccess:       true,
			expectWrittenAfterHijack:  true, // Hijack itself marks the writer as written
			expectError:               nil,
		},
		{
			name: "hijack after write should fail",
			action: func(w ResponseWriter) error {
				_, err := w.Write([]byte("test"))
				return err
			},
			expectWrittenBeforeHijack: true,
			expectHijackSuccess:       false,
			expectWrittenAfterHijack:  true,
			expectError:               errHijackAlreadyWritten,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hijacker := &mockHijacker{ResponseRecorder: httptest.NewRecorder()}
			writer := &responseWriter{}
			writer.reset(hijacker)
			w := ResponseWriter(writer)

			// Check initial state
			assert.False(t, w.Written(), "should not be written initially")

			// Perform pre-hijack action
			require.NoError(t, tc.action(w), "unexpected error during pre-hijack action")

			// Check state before hijacking
			assert.Equal(t, tc.expectWrittenBeforeHijack, w.Written(), "unexpected w.Written() state before hijack")

			// Attempt to hijack
			_, _, hijackErr := w.Hijack()

			// Check results
			require.ErrorIs(t, hijackErr, tc.expectError, "unexpected error from Hijack()")
			assert.Equal(t, tc.expectHijackSuccess, hijacker.hijacked, "unexpected hijacker.hijacked state")
			assert.Equal(t, tc.expectWrittenAfterHijack, w.Written(), "unexpected w.Written() state after hijack")
		})
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterHijackAfterWrite(t=%v)", t)
}

// Test: WebSocket compatibility - allow hijack after WriteHeaderNow(), but block after body data.
func TestResponseWriterHijackAfterWriteHeaderNow(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterHijackAfterWriteHeaderNow(t=%v)", t)
	tests := []struct {
		name                      string
		action                    func(w ResponseWriter) error
		expectWrittenBeforeHijack bool
		expectHijackSuccess       bool
		expectWrittenAfterHijack  bool
		expectError               error
	}{
		{
			name: "hijack after WriteHeaderNow only should succeed (websocket pattern)",
			action: func(w ResponseWriter) error {
				w.WriteHeaderNow() // Simulate websocket.Accept() behavior
				return nil
			},
			expectWrittenBeforeHijack: true,
			expectHijackSuccess:       true, // NEW BEHAVIOR: allow hijack after just header write
			expectWrittenAfterHijack:  true,
			expectError:               nil,
		},
		{
			name: "hijack after WriteHeaderNow + Write should fail",
			action: func(w ResponseWriter) error {
				w.WriteHeaderNow()
				_, err := w.Write([]byte("test"))
				return err
			},
			expectWrittenBeforeHijack: true,
			expectHijackSuccess:       false,
			expectWrittenAfterHijack:  true,
			expectError:               errHijackAlreadyWritten,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hijacker := &mockHijacker{ResponseRecorder: httptest.NewRecorder()}
			writer := &responseWriter{}
			writer.reset(hijacker)
			w := ResponseWriter(writer)

			require.NoError(t, tc.action(w), "unexpected error during pre-hijack action")

			assert.Equal(t, tc.expectWrittenBeforeHijack, w.Written(), "unexpected w.Written() state before hijack")

			_, _, hijackErr := w.Hijack()

			if tc.expectError == nil {
				require.NoError(t, hijackErr, "expected hijack to succeed")
			} else {
				require.ErrorIs(t, hijackErr, tc.expectError, "unexpected error from Hijack()")
			}
			assert.Equal(t, tc.expectHijackSuccess, hijacker.hijacked, "unexpected hijacker.hijacked state")
			assert.Equal(t, tc.expectWrittenAfterHijack, w.Written(), "unexpected w.Written() state after hijack")
		})
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterHijackAfterWriteHeaderNow(t=%v)", t)
}

func TestResponseWriterFlush(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterFlush(t=%v)", t)
	__logStart := logritetime.Now()
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &responseWriter{}
		writer.reset(w)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Flush()
	}))
	defer testServer.Close()

	// should return 500
	resp, err := http.Get(testServer.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	logritelog.Default().Printf("[ERROR] TestResponseWriterFlush(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterFlush(t=%v)", t)
}

func TestResponseWriterStatusCode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestResponseWriterStatusCode(t=%v)", t)
	testWriter := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(testWriter)
	w := ResponseWriter(writer)

	w.WriteHeader(http.StatusOK)
	w.WriteHeaderNow()

	assert.Equal(t, http.StatusOK, w.Status())
	assert.True(t, w.Written())

	w.WriteHeader(http.StatusUnauthorized)

	// status must be 200 although we tried to change it
	assert.Equal(t, http.StatusOK, w.Status())
	logritelog.Default().Printf("[WARN] <<< Exiting TestResponseWriterStatusCode(t=%v)", t)
}

// mockPusherResponseWriter is an http.ResponseWriter that implements http.Pusher.
type mockPusherResponseWriter struct {
	http.ResponseWriter
}

func (m *mockPusherResponseWriter) Push(target string, opts *http.PushOptions) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Push(target=%v,opts=%v)", target, opts)
	logritelog.Default().Printf("[WARN] <<< Exiting Push(target=%v,opts=%v)", target, opts)
	return nil
}

// nonPusherResponseWriter is an http.ResponseWriter that does not implement http.Pusher.
type nonPusherResponseWriter struct {
	http.ResponseWriter
}

func TestPusherWithPusher(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestPusherWithPusher(t=%v)", t)
	rw := &mockPusherResponseWriter{}
	w := &responseWriter{ResponseWriter: rw}

	pusher := w.Pusher()
	assert.NotNil(t, pusher, "Expected pusher to be non-nil")
	logritelog.Default().Printf("[WARN] <<< Exiting TestPusherWithPusher(t=%v)", t)
}

func TestPusherWithoutPusher(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestPusherWithoutPusher(t=%v)", t)
	rw := &nonPusherResponseWriter{}
	w := &responseWriter{ResponseWriter: rw}

	pusher := w.Pusher()
	assert.Nil(t, pusher, "Expected pusher to be nil")
	logritelog.Default().Printf("[WARN] <<< Exiting TestPusherWithoutPusher(t=%v)", t)
}
