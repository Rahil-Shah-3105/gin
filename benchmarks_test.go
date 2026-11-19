// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import (
	logritelog "log"
	logritetime "time"
)

import (
	"html/template"
	"net/http"
	"os"
	"testing"
)

func BenchmarkOneRoute(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkOneRoute(B=%v)", B)
	router := New()
	router.GET("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkOneRoute(B=%v)", B)
}

func BenchmarkRecoveryMiddleware(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkRecoveryMiddleware(B=%v)", B)
	router := New()
	router.Use(Recovery())
	router.GET("/", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkRecoveryMiddleware(B=%v)", B)
}

func BenchmarkLoggerMiddleware(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkLoggerMiddleware(B=%v)", B)
	router := New()
	router.Use(LoggerWithWriter(newMockWriter()))
	router.GET("/", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkLoggerMiddleware(B=%v)", B)
}

func BenchmarkManyHandlers(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkManyHandlers(B=%v)", B)
	router := New()
	router.Use(Recovery(), LoggerWithWriter(newMockWriter()))
	router.Use(func(c *Context) {})
	router.Use(func(c *Context) {})
	router.GET("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkManyHandlers(B=%v)", B)
}

func Benchmark5Params(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering Benchmark5Params(B=%v)", B)
	DefaultWriter = os.Stdout
	router := New()
	router.Use(func(c *Context) {})
	router.GET("/param/:param1/:params2/:param3/:param4/:param5", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/param/path/to/parameter/john/12345")
	logritelog.Default().Printf("[DEBUG] Benchmark5Params(B=%v): DefaultWriter → %v", B, DefaultWriter)
	logritelog.Default().Printf("[WARN] <<< Exiting Benchmark5Params(B=%v)", B)
}

func BenchmarkOneRouteJSON(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkOneRouteJSON(B=%v)", B)
	router := New()
	data := struct {
		Status string `json:"status"`
	}{"ok"}
	router.GET("/json", func(c *Context) {
		c.JSON(http.StatusOK, data)
	})
	runRequest(B, router, http.MethodGet, "/json")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkOneRouteJSON(B=%v)", B)
}

func BenchmarkOneRouteHTML(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkOneRouteHTML(B=%v)", B)
	router := New()
	t := template.Must(template.New("index").Parse(`
		<html><body><h1>{{.}}</h1></body></html>`))
	router.SetHTMLTemplate(t)

	router.GET("/html", func(c *Context) {
		c.HTML(http.StatusOK, "index", "hola")
	})
	runRequest(B, router, http.MethodGet, "/html")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkOneRouteHTML(B=%v)", B)
}

func BenchmarkOneRouteSet(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkOneRouteSet(B=%v)", B)
	router := New()
	router.GET("/ping", func(c *Context) {
		c.Set("key", "value")
	})
	runRequest(B, router, http.MethodGet, "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkOneRouteSet(B=%v)", B)
}

func BenchmarkOneRouteString(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkOneRouteString(B=%v)", B)
	router := New()
	router.GET("/text", func(c *Context) {
		c.String(http.StatusOK, "this is a plain text")
	})
	runRequest(B, router, http.MethodGet, "/text")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkOneRouteString(B=%v)", B)
}

func BenchmarkManyRoutesFirst(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkManyRoutesFirst(B=%v)", B)
	router := New()
	router.Any("/ping", func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkManyRoutesFirst(B=%v)", B)
}

func BenchmarkManyRoutesLast(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering BenchmarkManyRoutesLast(B=%v)", B)
	router := New()
	router.Any("/ping", func(c *Context) {})
	runRequest(B, router, "OPTIONS", "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting BenchmarkManyRoutesLast(B=%v)", B)
}

func Benchmark404(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering Benchmark404(B=%v)", B)
	router := New()
	router.Any("/something", func(c *Context) {})
	router.NoRoute(func(c *Context) {})
	runRequest(B, router, http.MethodGet, "/ping")
	logritelog.Default().Printf("[WARN] <<< Exiting Benchmark404(B=%v)", B)
}

func Benchmark404Many(B *testing.B) {
	logritelog.Default().Printf("[ERROR] >>> Entering Benchmark404Many(B=%v)", B)
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
	logritelog.Default().Printf("[WARN] <<< Exiting Benchmark404Many(B=%v)", B)
}

type mockWriter struct {
	headers http.Header
}

func newMockWriter() *mockWriter {
	logritelog.Default().Printf("[ERROR] >>> Entering newMockWriter()")
	return &mockWriter{
		http.Header{},
	}
}

func (m *mockWriter) Header() (h http.Header) {
	logritelog.Default().Printf("[ERROR] >>> Entering Header()")
	return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering Write(p=%v)", p)
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering WriteString(s=%v)", s)
	return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

func runRequest(B *testing.B, r *Engine, method, path string) {
	logritelog.Default().Printf("[ERROR] >>> Entering runRequest(B=%v,r=%v,method=%v,path=%v)", B, r, method, path)
	__logStart := logritetime.Now()
	// create fake request
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in runRequest(B=%v,r=%v,method=%v,path=%v): %v", B, r, method, path, err)
		panic(err)
	}
	w := newMockWriter()
	B.ReportAllocs()
	B.ResetTimer()
	for B.Loop() {
		r.ServeHTTP(w, req)
	}
	logritelog.Default().Printf("[ERROR] runRequest(B,r,method,path)=%v,%v,%v,%v: external service call took %v ms", B, r, method, path, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting runRequest(B=%v,r=%v,method=%v,path=%v)", B, r, method, path)
}
