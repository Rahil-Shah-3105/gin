// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import (
	logritelog "log"
	logritetime "time"
)

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func formatAsDate(t time.Time) string {
	logritelog.Default().Printf("[ERROR] >>> Entering formatAsDate(t=%v)", t)
	year, month, day := t.Date()
	logritelog.Default().Printf("[WARN] <<< Exiting formatAsDate(t=%v)", t)
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

func setupHTMLFiles(t *testing.T, mode string, tls bool, loadMethod func(*Engine)) *httptest.Server {
	logritelog.Default().Printf("[ERROR] >>> Entering setupHTMLFiles(t=%v,mode=%v,tls=%v,loadMethod=%v)", t, mode, tls, loadMethod)
	SetMode(mode)
	defer SetMode(TestMode)

	var router *Engine
	captureOutput(t, func() {
		router = New()
		router.Delims("{[{", "}]}")
		router.SetFuncMap(template.FuncMap{
			"formatAsDate": formatAsDate,
		})
		loadMethod(router)
		router.GET("/test", func(c *Context) {
			c.HTML(http.StatusOK, "hello.tmpl", map[string]string{"name": "world"})
		})
		router.GET("/raw", func(c *Context) {
			c.HTML(http.StatusOK, "raw.tmpl", map[string]any{
				"now": time.Date(2017, 07, 01, 0, 0, 0, 0, time.UTC), //nolint:gofumpt
			})
		})
	})

	var ts *httptest.Server

	if tls {
		ts = httptest.NewTLSServer(router)
	} else {
		ts = httptest.NewServer(router)
	}

	logritelog.Default().Printf("[WARN] <<< Exiting setupHTMLFiles(t=%v,mode=%v,tls=%v,loadMethod=%v)", t, mode, tls, loadMethod)
	return ts
}

func TestLoadHTMLGlobDebugMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLGlobDebugMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		DebugMode,
		false,
		func(router *Engine) {
			router.LoadHTMLGlob("./testdata/template/*")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestH2c(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestH2c(t=%v)", t)
	ln, err := net.Listen("tcp", localhostIP+":0")
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in TestH2c(t=%v): %v", t, err)
		t.Error(err)
	}
	r := Default()
	r.UseH2C = true
	logritelog.Default().Printf("[DEBUG] TestH2c(t=%v): r.UseH2C → %v", t, r.UseH2C)
	r.GET("/", func(c *Context) {
		c.String(200, "<h1>Hello world</h1>")
	})
	go func() {
		err := http.Serve(ln, r.Handler())
		if err != nil {
			t.Log(err)
		}
	}()
	defer ln.Close()

	url := "http://" + ln.Addr().String() + "/"

	httpClient := http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		},
	}

	res, err := httpClient.Get(url)
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
	logritelog.Default().Printf("[WARN] <<< Exiting TestH2c(t=%v)", t)
}

func TestLoadHTMLGlobTestMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLGlobTestMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		false,
		func(router *Engine) {
			router.LoadHTMLGlob("./testdata/template/*")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLGlobReleaseMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLGlobReleaseMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		ReleaseMode,
		false,
		func(router *Engine) {
			router.LoadHTMLGlob("./testdata/template/*")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLGlobUsingTLS(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLGlobUsingTLS(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		DebugMode,
		true,
		func(router *Engine) {
			router.LoadHTMLGlob("./testdata/template/*")
		},
	)
	defer ts.Close()

	// Use InsecureSkipVerify for avoiding `x509: certificate signed by unknown authority` error
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLGlobFromFuncMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLGlobFromFuncMap(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		DebugMode,
		false,
		func(router *Engine) {
			router.LoadHTMLGlob("./testdata/template/*")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/raw")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "Date: 2017/07/01", string(resp))
}

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	SetMode(TestMode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

func TestCreateEngine(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestCreateEngine(t=%v)", t)
	router := New()
	assert.Equal(t, "/", router.basePath)
	assert.Equal(t, router.engine, router)
	assert.Empty(t, router.Handlers)
	logritelog.Default().Printf("[WARN] <<< Exiting TestCreateEngine(t=%v)", t)
}

func TestLoadHTMLFilesTestMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFilesTestMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFiles("./testdata/template/hello.tmpl", "./testdata/template/raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFilesDebugMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFilesDebugMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		DebugMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFiles("./testdata/template/hello.tmpl", "./testdata/template/raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFilesReleaseMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFilesReleaseMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		ReleaseMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFiles("./testdata/template/hello.tmpl", "./testdata/template/raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFilesUsingTLS(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFilesUsingTLS(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		true,
		func(router *Engine) {
			router.LoadHTMLFiles("./testdata/template/hello.tmpl", "./testdata/template/raw.tmpl")
		},
	)
	defer ts.Close()

	// Use InsecureSkipVerify for avoiding `x509: certificate signed by unknown authority` error
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFilesFuncMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFilesFuncMap(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFiles("./testdata/template/hello.tmpl", "./testdata/template/raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/raw")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "Date: 2017/07/01", string(resp))
}

var tmplFS = http.Dir("testdata/template")

func TestLoadHTMLFSTestMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFSTestMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFS(tmplFS, "hello.tmpl", "raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFSDebugMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFSDebugMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		DebugMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFS(tmplFS, "hello.tmpl", "raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFSReleaseMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFSReleaseMode(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		ReleaseMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFS(tmplFS, "hello.tmpl", "raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFSUsingTLS(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFSUsingTLS(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		true,
		func(router *Engine) {
			router.LoadHTMLFS(tmplFS, "hello.tmpl", "raw.tmpl")
		},
	)
	defer ts.Close()

	// Use InsecureSkipVerify for avoiding `x509: certificate signed by unknown authority` error
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

func TestLoadHTMLFSFuncMap(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLoadHTMLFSFuncMap(t=%v)", t)
	ts := setupHTMLFiles(
		t,
		TestMode,
		false,
		func(router *Engine) {
			router.LoadHTMLFS(tmplFS, "hello.tmpl", "raw.tmpl")
		},
	)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/raw")
	if err != nil {
		t.Error(err)
	}

	resp, _ := io.ReadAll(res.Body)
	assert.Equal(t, "Date: 2017/07/01", string(resp))
}

func TestAddRoute(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestAddRoute(t=%v)", t)
	router := New()
	router.addRoute(http.MethodGet, "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, router.trees, 1)
	assert.NotNil(t, router.trees.get(http.MethodGet))
	assert.Nil(t, router.trees.get(http.MethodPost))

	router.addRoute(http.MethodPost, "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, router.trees, 2)
	assert.NotNil(t, router.trees.get(http.MethodGet))
	assert.NotNil(t, router.trees.get(http.MethodPost))

	router.addRoute(http.MethodPost, "/post", HandlersChain{func(_ *Context) {}})
	assert.Len(t, router.trees, 2)
	logritelog.Default().Printf("[WARN] <<< Exiting TestAddRoute(t=%v)", t)
}

func TestAddRouteFails(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestAddRouteFails(t=%v)", t)
	router := New()
	assert.Panics(t, func() { router.addRoute("", "/", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { router.addRoute(http.MethodGet, "a", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { router.addRoute(http.MethodGet, "/", HandlersChain{}) })

	router.addRoute(http.MethodPost, "/post", HandlersChain{func(_ *Context) {}})
	assert.Panics(t, func() {
		router.addRoute(http.MethodPost, "/post", HandlersChain{func(_ *Context) {}})
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestAddRouteFails(t=%v)", t)
}

func TestCreateDefaultRouter(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestCreateDefaultRouter(t=%v)", t)
	router := Default()
	assert.Len(t, router.Handlers, 2)
	logritelog.Default().Printf("[WARN] <<< Exiting TestCreateDefaultRouter(t=%v)", t)
}

func TestNoRouteWithoutGlobalHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestNoRouteWithoutGlobalHandlers(t=%v)", t)
	var middleware0 HandlerFunc = func(c *Context) {}
	var middleware1 HandlerFunc = func(c *Context) {}

	router := New()

	router.NoRoute(middleware0)
	assert.Nil(t, router.Handlers)
	assert.Len(t, router.noRoute, 1)
	assert.Len(t, router.allNoRoute, 1)
	compareFunc(t, router.noRoute[0], middleware0)
	compareFunc(t, router.allNoRoute[0], middleware0)

	router.NoRoute(middleware1, middleware0)
	assert.Len(t, router.noRoute, 2)
	assert.Len(t, router.allNoRoute, 2)
	compareFunc(t, router.noRoute[0], middleware1)
	compareFunc(t, router.allNoRoute[0], middleware1)
	compareFunc(t, router.noRoute[1], middleware0)
	compareFunc(t, router.allNoRoute[1], middleware0)
	logritelog.Default().Printf("[WARN] <<< Exiting TestNoRouteWithoutGlobalHandlers(t=%v)", t)
}

func TestNoRouteWithGlobalHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestNoRouteWithGlobalHandlers(t=%v)", t)
	var middleware0 HandlerFunc = func(c *Context) {}
	var middleware1 HandlerFunc = func(c *Context) {}
	var middleware2 HandlerFunc = func(c *Context) {}

	router := New()
	router.Use(middleware2)

	router.NoRoute(middleware0)
	assert.Len(t, router.allNoRoute, 2)
	assert.Len(t, router.Handlers, 1)
	assert.Len(t, router.noRoute, 1)

	compareFunc(t, router.Handlers[0], middleware2)
	compareFunc(t, router.noRoute[0], middleware0)
	compareFunc(t, router.allNoRoute[0], middleware2)
	compareFunc(t, router.allNoRoute[1], middleware0)

	router.Use(middleware1)
	assert.Len(t, router.allNoRoute, 3)
	assert.Len(t, router.Handlers, 2)
	assert.Len(t, router.noRoute, 1)

	compareFunc(t, router.Handlers[0], middleware2)
	compareFunc(t, router.Handlers[1], middleware1)
	compareFunc(t, router.noRoute[0], middleware0)
	compareFunc(t, router.allNoRoute[0], middleware2)
	compareFunc(t, router.allNoRoute[1], middleware1)
	compareFunc(t, router.allNoRoute[2], middleware0)
	logritelog.Default().Printf("[WARN] <<< Exiting TestNoRouteWithGlobalHandlers(t=%v)", t)
}

func TestNoMethodWithoutGlobalHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestNoMethodWithoutGlobalHandlers(t=%v)", t)
	var middleware0 HandlerFunc = func(c *Context) {}
	var middleware1 HandlerFunc = func(c *Context) {}

	router := New()

	router.NoMethod(middleware0)
	assert.Empty(t, router.Handlers)
	assert.Len(t, router.noMethod, 1)
	assert.Len(t, router.allNoMethod, 1)
	compareFunc(t, router.noMethod[0], middleware0)
	compareFunc(t, router.allNoMethod[0], middleware0)

	router.NoMethod(middleware1, middleware0)
	assert.Len(t, router.noMethod, 2)
	assert.Len(t, router.allNoMethod, 2)
	compareFunc(t, router.noMethod[0], middleware1)
	compareFunc(t, router.allNoMethod[0], middleware1)
	compareFunc(t, router.noMethod[1], middleware0)
	compareFunc(t, router.allNoMethod[1], middleware0)
	logritelog.Default().Printf("[WARN] <<< Exiting TestNoMethodWithoutGlobalHandlers(t=%v)", t)
}

func TestRebuild404Handlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRebuild404Handlers(t=%v)", t)
	var middleware0 HandlerFunc = func(c *Context) {}
	var middleware1 HandlerFunc = func(c *Context) {}

	router := New()

	// Initially, allNoRoute should be nil
	assert.Nil(t, router.allNoRoute)

	// Set NoRoute handlers
	router.NoRoute(middleware0)
	assert.Len(t, router.allNoRoute, 1)
	assert.Len(t, router.noRoute, 1)
	compareFunc(t, router.allNoRoute[0], middleware0)

	// Add Use middleware should trigger rebuild404Handlers
	router.Use(middleware1)
	assert.Len(t, router.allNoRoute, 2)
	assert.Len(t, router.Handlers, 1)
	assert.Len(t, router.noRoute, 1)

	// Global middleware should come first
	compareFunc(t, router.allNoRoute[0], middleware1)
	compareFunc(t, router.allNoRoute[1], middleware0)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRebuild404Handlers(t=%v)", t)
}

func TestNoMethodWithGlobalHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestNoMethodWithGlobalHandlers(t=%v)", t)
	var middleware0 HandlerFunc = func(c *Context) {}
	var middleware1 HandlerFunc = func(c *Context) {}
	var middleware2 HandlerFunc = func(c *Context) {}

	router := New()
	router.Use(middleware2)

	router.NoMethod(middleware0)
	assert.Len(t, router.allNoMethod, 2)
	assert.Len(t, router.Handlers, 1)
	assert.Len(t, router.noMethod, 1)

	compareFunc(t, router.Handlers[0], middleware2)
	compareFunc(t, router.noMethod[0], middleware0)
	compareFunc(t, router.allNoMethod[0], middleware2)
	compareFunc(t, router.allNoMethod[1], middleware0)

	router.Use(middleware1)
	assert.Len(t, router.allNoMethod, 3)
	assert.Len(t, router.Handlers, 2)
	assert.Len(t, router.noMethod, 1)

	compareFunc(t, router.Handlers[0], middleware2)
	compareFunc(t, router.Handlers[1], middleware1)
	compareFunc(t, router.noMethod[0], middleware0)
	compareFunc(t, router.allNoMethod[0], middleware2)
	compareFunc(t, router.allNoMethod[1], middleware1)
	compareFunc(t, router.allNoMethod[2], middleware0)
	logritelog.Default().Printf("[WARN] <<< Exiting TestNoMethodWithGlobalHandlers(t=%v)", t)
}

func compareFunc(t *testing.T, a, b any) {
	logritelog.Default().Printf("[ERROR] >>> Entering compareFunc(t=%v,a=%v,b=%v)", t, a, b)
	sf1 := reflect.ValueOf(a)
	sf2 := reflect.ValueOf(b)
	if sf1.Pointer() != sf2.Pointer() {
		t.Error("different functions")
	}
	logritelog.Default().Printf("[WARN] <<< Exiting compareFunc(t=%v,a=%v,b=%v)", t, a, b)
}

func TestListOfRoutes(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestListOfRoutes(t=%v)", t)
	router := New()
	router.GET("/favicon.ico", handlerTest1)
	router.GET("/", handlerTest1)
	group := router.Group("/users")
	{
		group.GET("/", handlerTest2)
		group.GET("/:id", handlerTest1)
		group.POST("/:id", handlerTest2)
	}
	router.Static("/static", ".")

	list := router.Routes()

	assert.Len(t, list, 7)
	assertRoutePresent(t, list, RouteInfo{
		Method:  http.MethodGet,
		Path:    "/favicon.ico",
		Handler: "^(.*/vendor/)?github.com/gin-gonic/gin.handlerTest1$",
	})
	assertRoutePresent(t, list, RouteInfo{
		Method:  http.MethodGet,
		Path:    "/",
		Handler: "^(.*/vendor/)?github.com/gin-gonic/gin.handlerTest1$",
	})
	assertRoutePresent(t, list, RouteInfo{
		Method:  http.MethodGet,
		Path:    "/users/",
		Handler: "^(.*/vendor/)?github.com/gin-gonic/gin.handlerTest2$",
	})
	assertRoutePresent(t, list, RouteInfo{
		Method:  http.MethodGet,
		Path:    "/users/:id",
		Handler: "^(.*/vendor/)?github.com/gin-gonic/gin.handlerTest1$",
	})
	assertRoutePresent(t, list, RouteInfo{
		Method:  http.MethodPost,
		Path:    "/users/:id",
		Handler: "^(.*/vendor/)?github.com/gin-gonic/gin.handlerTest2$",
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestListOfRoutes(t=%v)", t)
}

func TestEngineHandleContext(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContext(t=%v)", t)
	r := New()
	r.GET("/", func(c *Context) {
		c.Request.URL.Path = "/v2"
		r.HandleContext(c)
	})
	v2 := r.Group("/v2")
	{
		v2.GET("/", func(c *Context) {})
	}

	assert.NotPanics(t, func() {
		w := PerformRequest(r, http.MethodGet, "/")
		assert.Equal(t, 301, w.Code)
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContext(t=%v)", t)
}

func TestEngineHandleContextManyReEntries(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextManyReEntries(t=%v)", t)
	expectValue := 10000

	var handlerCounter, middlewareCounter int64

	r := New()
	r.Use(func(c *Context) {
		atomic.AddInt64(&middlewareCounter, 1)
	})
	r.GET("/:count", func(c *Context) {
		countStr := c.Param("count")
		count, err := strconv.Atoi(countStr)
		require.NoError(t, err)

		n, err := c.Writer.Write([]byte("."))
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		switch {
		case count > 0:
			c.Request.URL.Path = "/" + strconv.Itoa(count-1)
			r.HandleContext(c)
		}
	}, func(c *Context) {
		atomic.AddInt64(&handlerCounter, 1)
	})

	assert.NotPanics(t, func() {
		w := PerformRequest(r, http.MethodGet, "/"+strconv.Itoa(expectValue-1)) // include 0 value
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, expectValue, w.Body.Len())
	})

	assert.Equal(t, int64(expectValue), handlerCounter)
	assert.Equal(t, int64(expectValue), middlewareCounter)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextManyReEntries(t=%v)", t)
}

func TestEngineHandleContextPreventsMiddlewareReEntry(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextPreventsMiddlewareReEntry(t=%v)", t)
	// given
	var handlerCounterV1, handlerCounterV2, middlewareCounterV1 int64

	r := New()
	v1 := r.Group("/v1")
	{
		v1.Use(func(c *Context) {
			atomic.AddInt64(&middlewareCounterV1, 1)
		})
		v1.GET("/test", func(c *Context) {
			atomic.AddInt64(&handlerCounterV1, 1)
			c.Status(http.StatusOK)
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/test", func(c *Context) {
			c.Request.URL.Path = "/v1/test"
			r.HandleContext(c)
		}, func(c *Context) {
			atomic.AddInt64(&handlerCounterV2, 1)
		})
	}

	// when
	responseV1 := PerformRequest(r, "GET", "/v1/test")
	responseV2 := PerformRequest(r, "GET", "/v2/test")

	// then
	assert.Equal(t, 200, responseV1.Code)
	assert.Equal(t, 200, responseV2.Code)
	assert.Equal(t, int64(2), handlerCounterV1)
	assert.Equal(t, int64(2), middlewareCounterV1)
	assert.Equal(t, int64(1), handlerCounterV2)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextPreventsMiddlewareReEntry(t=%v)", t)
}

func TestEngineHandleContextNoRouteWithGroupMiddleware(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextNoRouteWithGroupMiddleware(t=%v)", t)
	// Scenario from issue #1848:
	// - Engine with no global middleware (gin.New())
	// - A group with middleware
	// - A route in that group
	// - NoRoute handler that redirects via HandleContext
	// The group middleware should run exactly once per HandleContext call,
	// not accumulate across redirects.

	var middlewareCount, handlerCount int64

	r := New()
	grp := r.Group("", func(c *Context) {
		atomic.AddInt64(&middlewareCount, 1)
		c.Next()
	})
	grp.GET("/target", func(c *Context) {
		atomic.AddInt64(&handlerCount, 1)
		c.String(http.StatusOK, "ok")
	})

	r.NoRoute(func(c *Context) {
		c.Request.URL.Path = "/target"
		r.HandleContext(c)
	})

	// Access a non-existent route to trigger NoRoute -> HandleContext
	w := PerformRequest(r, "GET", "/nonexistent")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	// Middleware and handler should each run exactly once
	assert.Equal(t, int64(1), atomic.LoadInt64(&middlewareCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&handlerCount))
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextNoRouteWithGroupMiddleware(t=%v)", t)
}

func TestEngineHandleContextNoRouteWithEngineMiddleware(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextNoRouteWithEngineMiddleware(t=%v)", t)
	// When engine middleware exists and NoRoute redirects via HandleContext,
	// verify the handlers run the expected number of times.

	var engineMwCount, groupMwCount, handlerCount int64

	r := New()
	r.Use(func(c *Context) {
		atomic.AddInt64(&engineMwCount, 1)
		c.Next()
	})

	grp := r.Group("", func(c *Context) {
		atomic.AddInt64(&groupMwCount, 1)
		c.Next()
	})
	grp.GET("/target", func(c *Context) {
		atomic.AddInt64(&handlerCount, 1)
		c.String(http.StatusOK, "ok")
	})

	r.NoRoute(func(c *Context) {
		c.Request.URL.Path = "/target"
		r.HandleContext(c)
	})

	w := PerformRequest(r, "GET", "/nonexistent")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	// Handler and group middleware should each run once (from HandleContext)
	assert.Equal(t, int64(1), atomic.LoadInt64(&handlerCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&groupMwCount))
	// Engine middleware runs twice: once for the NoRoute chain, once for the HandleContext chain
	// This is expected behavior since HandleContext re-enters the full handler chain
	assert.Equal(t, int64(2), atomic.LoadInt64(&engineMwCount))
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextNoRouteWithEngineMiddleware(t=%v)", t)
}

func TestEngineHandleContextUseEscapedPathPercentEncoded(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextUseEscapedPathPercentEncoded(t=%v)", t)
	r := New()
	r.UseEscapedPath = true
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseEscapedPathPercentEncoded(t=%v): r.UseEscapedPath → %v", t, r.UseEscapedPath)
	r.UnescapePathValues = false
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseEscapedPathPercentEncoded(t=%v): r.UnescapePathValues → %v", t, r.UnescapePathValues)

	r.GET("/v1/:path", func(c *Context) {
		// Path is Escaped, the %25 is not interpreted as %
		assert.Equal(t, "foo%252Fbar", c.Param("path"))
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/foo%252Fbar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextUseEscapedPathPercentEncoded(t=%v)", t)
}

func TestEngineHandleContextUseRawPathPercentEncoded(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextUseRawPathPercentEncoded(t=%v)", t)
	r := New()
	r.UseRawPath = true
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseRawPathPercentEncoded(t=%v): r.UseRawPath → %v", t, r.UseRawPath)
	r.UnescapePathValues = false
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseRawPathPercentEncoded(t=%v): r.UnescapePathValues → %v", t, r.UnescapePathValues)

	r.GET("/v1/:path", func(c *Context) {
		// Path is used, the %25 is interpreted as %
		assert.Equal(t, "foo%2Fbar", c.Param("path"))
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/foo%252Fbar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextUseRawPathPercentEncoded(t=%v)", t)
}

func TestEngineHandleContextUseEscapedPathOverride(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEngineHandleContextUseEscapedPathOverride(t=%v)", t)
	r := New()
	r.UseEscapedPath = true
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseEscapedPathOverride(t=%v): r.UseEscapedPath → %v", t, r.UseEscapedPath)
	r.UseRawPath = true
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseEscapedPathOverride(t=%v): r.UseRawPath → %v", t, r.UseRawPath)
	r.UnescapePathValues = false
	logritelog.Default().Printf("[DEBUG] TestEngineHandleContextUseEscapedPathOverride(t=%v): r.UnescapePathValues → %v", t, r.UnescapePathValues)

	r.GET("/v1/:path", func(c *Context) {
		assert.Equal(t, "foo%25bar", c.Param("path"))
		c.Status(http.StatusOK)
	})

	assert.NotPanics(t, func() {
		w := PerformRequest(r, http.MethodGet, "/v1/foo%25bar")
		assert.Equal(t, 200, w.Code)
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestEngineHandleContextUseEscapedPathOverride(t=%v)", t)
}

func TestPrepareTrustedCIRDsWith(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestPrepareTrustedCIRDsWith(t=%v)", t)
	r := New()

	// valid ipv4 cidr
	{
		expectedTrustedCIDRs := []*net.IPNet{parseCIDR("0.0.0.0/0")}
		err := r.SetTrustedProxies([]string{"0.0.0.0/0"})

		require.NoError(t, err)
		assert.Equal(t, expectedTrustedCIDRs, r.trustedCIDRs)
	}

	// invalid ipv4 cidr
	{
		err := r.SetTrustedProxies([]string{"192.168.1.33/33"})

		require.Error(t, err)
	}

	// valid ipv4 address
	{
		expectedTrustedCIDRs := []*net.IPNet{parseCIDR("192.168.1.33/32")}

		err := r.SetTrustedProxies([]string{"192.168.1.33"})

		require.NoError(t, err)
		assert.Equal(t, expectedTrustedCIDRs, r.trustedCIDRs)
	}

	// invalid ipv4 address
	{
		err := r.SetTrustedProxies([]string{"192.168.1.256"})

		require.Error(t, err)
	}

	// valid ipv6 address
	{
		expectedTrustedCIDRs := []*net.IPNet{parseCIDR("2002:0000:0000:1234:abcd:ffff:c0a8:0101/128")}
		err := r.SetTrustedProxies([]string{"2002:0000:0000:1234:abcd:ffff:c0a8:0101"})

		require.NoError(t, err)
		assert.Equal(t, expectedTrustedCIDRs, r.trustedCIDRs)
	}

	// invalid ipv6 address
	{
		err := r.SetTrustedProxies([]string{"gggg:0000:0000:1234:abcd:ffff:c0a8:0101"})

		require.Error(t, err)
	}

	// valid ipv6 cidr
	{
		expectedTrustedCIDRs := []*net.IPNet{parseCIDR("::/0")}
		err := r.SetTrustedProxies([]string{"::/0"})

		require.NoError(t, err)
		assert.Equal(t, expectedTrustedCIDRs, r.trustedCIDRs)
	}

	// invalid ipv6 cidr
	{
		err := r.SetTrustedProxies([]string{"gggg:0000:0000:1234:abcd:ffff:c0a8:0101/129"})

		require.Error(t, err)
	}

	// valid combination
	{
		expectedTrustedCIDRs := []*net.IPNet{
			parseCIDR("::/0"),
			parseCIDR("192.168.0.0/16"),
			parseCIDR("172.16.0.1/32"),
		}
		err := r.SetTrustedProxies([]string{
			"::/0",
			"192.168.0.0/16",
			"172.16.0.1",
		})

		require.NoError(t, err)
		assert.Equal(t, expectedTrustedCIDRs, r.trustedCIDRs)
	}

	// invalid combination
	{
		err := r.SetTrustedProxies([]string{
			"::/0",
			"192.168.0.0/16",
			"172.16.0.256",
		})

		require.Error(t, err)
	}

	// nil value
	{
		err := r.SetTrustedProxies(nil)

		assert.Nil(t, r.trustedCIDRs)
		require.NoError(t, err)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting TestPrepareTrustedCIRDsWith(t=%v)", t)
}

func parseCIDR(cidr string) *net.IPNet {
	logritelog.Default().Printf("[ERROR] >>> Entering parseCIDR(cidr=%v)", cidr)
	_, parsedCIDR, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println(err)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting parseCIDR(cidr=%v)", cidr)
	return parsedCIDR
}

func assertRoutePresent(t *testing.T, gotRoutes RoutesInfo, wantRoute RouteInfo) {
	logritelog.Default().Printf("[ERROR] >>> Entering assertRoutePresent(t=%v,gotRoutes=%v,wantRoute=%v)", t, gotRoutes, wantRoute)
	for _, gotRoute := range gotRoutes {
		if gotRoute.Path == wantRoute.Path && gotRoute.Method == wantRoute.Method {
			assert.Regexp(t, wantRoute.Handler, gotRoute.Handler)
			return
		}
	}
	t.Errorf("route not found: %v", wantRoute)
	logritelog.Default().Printf("[WARN] <<< Exiting assertRoutePresent(t=%v,gotRoutes=%v,wantRoute=%v)", t, gotRoutes, wantRoute)
}

func handlerTest1(c *Context) {}
func handlerTest2(c *Context) {}

func TestNewOptionFunc(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestNewOptionFunc(t=%v)", t)
	fc := func(e *Engine) {
		e.GET("/test1", handlerTest1)
		e.GET("/test2", handlerTest2)

		e.Use(func(c *Context) {
			c.Next()
		})
	}

	r := New(fc)

	routes := r.Routes()
	assertRoutePresent(t, routes, RouteInfo{Path: "/test1", Method: http.MethodGet, Handler: "github.com/gin-gonic/gin.handlerTest1"})
	assertRoutePresent(t, routes, RouteInfo{Path: "/test2", Method: http.MethodGet, Handler: "github.com/gin-gonic/gin.handlerTest2"})
	logritelog.Default().Printf("[WARN] <<< Exiting TestNewOptionFunc(t=%v)", t)
}

func TestWithOptionFunc(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestWithOptionFunc(t=%v)", t)
	r := New()

	r.With(func(e *Engine) {
		e.GET("/test1", handlerTest1)
		e.GET("/test2", handlerTest2)

		e.Use(func(c *Context) {
			c.Next()
		})
	})

	routes := r.Routes()
	assertRoutePresent(t, routes, RouteInfo{Path: "/test1", Method: http.MethodGet, Handler: "github.com/gin-gonic/gin.handlerTest1"})
	assertRoutePresent(t, routes, RouteInfo{Path: "/test2", Method: http.MethodGet, Handler: "github.com/gin-gonic/gin.handlerTest2"})
	logritelog.Default().Printf("[WARN] <<< Exiting TestWithOptionFunc(t=%v)", t)
}

type Birthday string

func (b *Birthday) UnmarshalParam(param string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering UnmarshalParam(param=%v)", param)
	*b = Birthday(strings.ReplaceAll(param, "-", "/"))
	logritelog.Default().Printf("[WARN] <<< Exiting UnmarshalParam(param=%v)", param)
	return nil
}

func TestCustomUnmarshalStruct(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestCustomUnmarshalStruct(t=%v)", t)
	route := Default()
	var request struct {
		Birthday Birthday `form:"birthday"`
	}
	route.GET("/test", func(ctx *Context) {
		_ = ctx.BindQuery(&request)
		ctx.JSON(200, request.Birthday)
	})
	req := httptest.NewRequest(http.MethodGet, "/test?birthday=2000-01-01", nil)
	w := httptest.NewRecorder()
	route.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `"2000/01/01"`, w.Body.String())
	logritelog.Default().Printf("[WARN] <<< Exiting TestCustomUnmarshalStruct(t=%v)", t)
}

// Test the fix for https://github.com/gin-gonic/gin/issues/4002
func TestMethodNotAllowedNoRoute(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMethodNotAllowedNoRoute(t=%v)", t)
	g := New()
	g.HandleMethodNotAllowed = true
	logritelog.Default().Printf("[DEBUG] TestMethodNotAllowedNoRoute(t=%v): g.HandleMethodNotAllowed → %v", t, g.HandleMethodNotAllowed)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	assert.NotPanics(t, func() { g.ServeHTTP(resp, req) })
	assert.Equal(t, http.StatusNotFound, resp.Code)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMethodNotAllowedNoRoute(t=%v)", t)
}

// Test the fix for https://github.com/gin-gonic/gin/pull/4415
func TestLiteralColonWithRun(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLiteralColonWithRun(t=%v)", t)
	__logStart := logritetime.Now()
	SetMode(TestMode)
	router := New()

	router.GET(`/test\:action`, func(c *Context) {
		c.JSON(http.StatusOK, H{"path": "literal_colon"})
	})

	router.updateRouteTrees()

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/test:action", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "literal_colon")
	logritelog.Default().Printf("[ERROR] TestLiteralColonWithRun(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestLiteralColonWithRun(t=%v)", t)
}

func TestLiteralColonWithDirectServeHTTP(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLiteralColonWithDirectServeHTTP(t=%v)", t)
	SetMode(TestMode)
	router := New()

	router.GET(`/test\:action`, func(c *Context) {
		c.JSON(http.StatusOK, H{"path": "literal_colon"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test:action", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "literal_colon")
	logritelog.Default().Printf("[ERROR] TestLiteralColonWithDirectServeHTTP(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestLiteralColonWithDirectServeHTTP(t=%v)", t)
}

func TestLiteralColonWithHandler(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLiteralColonWithHandler(t=%v)", t)
	SetMode(TestMode)
	router := New()

	router.GET(`/test\:action`, func(c *Context) {
		c.JSON(http.StatusOK, H{"path": "literal_colon"})
	})

	handler := router.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test:action", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "literal_colon")
	logritelog.Default().Printf("[ERROR] TestLiteralColonWithHandler(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestLiteralColonWithHandler(t=%v)", t)
}

func TestLiteralColonWithHTTPServer(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestLiteralColonWithHTTPServer(t=%v)", t)
	SetMode(TestMode)
	router := New()

	router.GET(`/test\:action`, func(c *Context) {
		c.JSON(http.StatusOK, H{"path": "literal_colon"})
	})

	router.GET("/test/:param", func(c *Context) {
		c.JSON(http.StatusOK, H{"param": c.Param("param")})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test:action", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "literal_colon")

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test/foo", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "foo")
	logritelog.Default().Printf("[ERROR] TestLiteralColonWithHTTPServer(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestLiteralColonWithHTTPServer(t=%v)", t)
}

// Test that updateRouteTrees is called only once
func TestUpdateRouteTreesCalledOnce(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestUpdateRouteTreesCalledOnce(t=%v)", t)
	SetMode(TestMode)
	router := New()

	router.GET(`/test\:action`, func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	for range 5 {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test:action", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	}
	logritelog.Default().Printf("[ERROR] TestUpdateRouteTreesCalledOnce(t)=%v: external service call took %v ms", t, logritetime.Since(__logStart).Milliseconds())
logritelog.Default().Printf("[WARN] <<< Exiting TestUpdateRouteTreesCalledOnce(t=%v)", t)
}
