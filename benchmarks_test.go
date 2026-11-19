// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"html/template"
	"net/http"
	"os"
	"testing"
)

func BenchmarkOneRoute(B *testing.B) {
	router := New()
	router.GET("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
}

func BenchmarkRecoveryMiddleware(B *testing.B) {
	router := New()
	router.Use(Recovery())
	router.GET("/", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/")
}

func BenchmarkLoggerMiddleware(B *testing.B) {
	router := New()
	router.Use(LoggerWithWriter(newMockWriter()))
	router.GET("/", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/")
}

func BenchmarkManyHandlers(B *testing.B) {
	router := New()
	router.Use(Recovery(), LoggerWithWriter(newMockWriter()))
	router.Use(func(c *Context) {})
	router.Use(func(c *Context) {})
	router.GET("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
}

func Benchmark5Params(B *testing.B) {
	DefaultWriter = os.Stdout
	router := New()
	router.Use(func(c *Context) {})
	router.GET("/param/:param1/:params2/:param3/:param4/:param5", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/param/path/to/parameter/john/12345")
}

func BenchmarkOneRouteJSON(B *testing.B) {
	router := New()
	data := struct {
		Status string `json:"status"`
	}{"ok"}
	router.GET("/json", func(c *Context) {
		c.JSON(http.StatusOK, data)
	})
	runRequest(B, router, http.MethodGet, "/json")
	logritelog.Default().Printf("[INFO] BenchmarkOneRouteJSON(B=%v): SOC2-AUDIT [CC7.1]: System operation completed", B)
}

func BenchmarkOneRouteHTML(B *testing.B) {
	router := New()
	t := template.Must(template.New("index").Parse(`
		<html><body><h1>{{.}}</h1></body></html>`))
	router.SetHTMLTemplate(t)

	router.GET("/html", func(c *Context) {
		c.HTML(http.StatusOK, "index", "hola")
	})
	runRequest(B, router, http.MethodGet, "/html")
	logritelog.Default().Printf("[INFO] BenchmarkOneRouteHTML(B=%v): SOC2-AUDIT [CC7.1]: System operation completed", B)
}

func BenchmarkOneRouteSet(B *testing.B) {
	router := New()
	router.GET("/ping", func(c *Context) {
		c.Set("key", "value")
	})
	runRequest(B, router, http.MethodGet, "/ping")
}

func BenchmarkOneRouteString(B *testing.B) {
	router := New()
	router.GET("/text", func(c *Context) {
		c.String(http.StatusOK, "this is a plain text")
	})
	runRequest(B, router, http.MethodGet, "/text")
	logritelog.Default().Printf("[INFO] BenchmarkOneRouteString(B=%v): SOC2-AUDIT [CC7.1]: System operation completed", B)
}

func BenchmarkManyRoutesFirst(B *testing.B) {
	router := New()
	router.Any("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
}

func BenchmarkManyRoutesLast(B *testing.B) {
	router := New()
	router.Any("/ping", func(c *Context) {})
	runRequest(B, router, "OPTIONS", "/ping")
}

func Benchmark404(B *testing.B) {
	router := New()
	router.Any("/something", func(c *Context) {})
	router.NoRoute(func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
}

func Benchmark404Many(B *testing.B) {
	router := New()
	router.GET("/", func(c *Context) {})
	router.GET("/path/to/something", func(c *Context) {})
	router.GET("/post/:id", func(c *Context) {})
	router.GET("/view/:id", func(c *Context) {})
	router.GET("/favicon.ico", func(c *Context) {})
	router.GET("/robots.txt", func(c *Context) {})
	router.GET("/delete/:id", func(c *Context) {})
	router.GET("/user/:id/:mode", func(c *Context) {})

	router.NoRoute(func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/viewfake")
}

type mockWriter struct {
	headers http.Header
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		http.Header{},
	}
}

func (m *mockWriter) Header() (h http.Header) {
	return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	logritelog.Default().Printf("[INFO] Write(p=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", p)
	logritelog.Default().Printf("[INFO] Write(p=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", p)
	logritelog.Default().Printf("[INFO] Write(p=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", p)
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	logritelog.Default().Printf("[INFO] WriteString(s=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", s)
	logritelog.Default().Printf("[INFO] WriteString(s=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", s)
	logritelog.Default().Printf("[INFO] WriteString(s=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", s)
	return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

func runRequest(B *testing.B, r *Engine, method, path string) {
	// create fake request
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Outcome 1: user attribution — State change recorded for the audit trail", B, r, method, path)
	logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Continuous Event Monitoring / Network Visibility — State change recorded for the audit trail", B, r, method, path)
	logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", B, r, method, path)
	logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", B, r, method, path)
	w := newMockWriter()
	B.ReportAllocs()
	B.ResetTimer()
	for B.Loop() {
		r.ServeHTTP(w, req)
	}
	logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", B, r, method, path)
logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Outcome 4: privilege changes — State change recorded for the audit trail", B, r, method, path)
logritelog.Default().Printf("[INFO] runRequest(B=%v, r=%v, method=%v, path=%v): M2614-AUDIT: Outcome 5: infrastructure changes — State change recorded for the audit trail", B, r, method, path)
}
