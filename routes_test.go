// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type header struct {
	Key   string
	Value string
}

// PerformRequest for testing gin router.
func PerformRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	logritelog.Default().Printf(">>> Entering PerformRequest(r=%v,method=%v,path=%v,headers=%v)", r, method, path, headers)
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	logritelog.Default().Printf("<<< Exiting PerformRequest(r=%v,method=%v,path=%v,headers=%v)", r, method, path, headers)
	return w
}

func testRouteOK(method string, t *testing.T) {
	logritelog.Default().Printf(">>> Entering testRouteOK(method=%v,t=%v)", method, t)
	passed := false
	passedAny := false
	r := New()
	r.Any("/test2", func(c *Context) {
		passedAny = true
		logritelog.Default().Printf("testRouteOK(method=%v,t=%v): passedAny → %v", method, t, passedAny)
	})
	r.Handle(method, "/test", func(c *Context) {
		passed = true
		logritelog.Default().Printf("testRouteOK(method=%v,t=%v): passed → %v", method, t, passed)
	})

	w := PerformRequest(r, method, "/test")
	assert.True(t, passed)
	assert.Equal(t, http.StatusOK, w.Code)

	PerformRequest(r, method, "/test2")
	assert.True(t, passedAny)
	logritelog.Default().Printf("<<< Exiting testRouteOK(method=%v,t=%v)", method, t)
}

// TestSingleRouteOK tests that POST route is correctly invoked.
func testRouteNotOK(method string, t *testing.T) {
	logritelog.Default().Printf(">>> Entering testRouteNotOK(method=%v,t=%v)", method, t)
	passed := false
	router := New()
	router.Handle(method, "/test_2", func(c *Context) {
		passed = true
		logritelog.Default().Printf("testRouteNotOK(method=%v,t=%v): passed → %v", method, t, passed)
	})

	w := PerformRequest(router, method, "/test")

	assert.False(t, passed)
	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("<<< Exiting testRouteNotOK(method=%v,t=%v)", method, t)
}

// TestSingleRouteOK tests that POST route is correctly invoked.
func testRouteNotOK2(method string, t *testing.T) {
	logritelog.Default().Printf(">>> Entering testRouteNotOK2(method=%v,t=%v)", method, t)
	passed := false
	router := New()
	router.HandleMethodNotAllowed = true
	logritelog.Default().Printf("testRouteNotOK2(method=%v,t=%v): router.HandleMethodNotAllowed → %v", method, t, router.HandleMethodNotAllowed)
	var methodRoute string
	if method == http.MethodPost {
		methodRoute = http.MethodGet
		logritelog.Default().Printf("testRouteNotOK2(method=%v,t=%v): methodRoute → %v", method, t, methodRoute)
	} else {
		methodRoute = http.MethodPost
	}
	router.Handle(methodRoute, "/test", func(c *Context) {
		passed = true
		logritelog.Default().Printf("testRouteNotOK2(method=%v,t=%v): passed → %v", method, t, passed)
	})

	w := PerformRequest(router, method, "/test")

	assert.False(t, passed)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	logritelog.Default().Printf("<<< Exiting testRouteNotOK2(method=%v,t=%v)", method, t)
}

func TestRouterMethod(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterMethod(t=%v)", t)
	router := New()
	router.PUT("/hey2", func(c *Context) {
		c.String(http.StatusOK, "sup2")
	})

	router.PUT("/hey", func(c *Context) {
		c.String(http.StatusOK, "called")
	})

	router.PUT("/hey3", func(c *Context) {
		c.String(http.StatusOK, "sup3")
	})

	w := PerformRequest(router, http.MethodPut, "/hey")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "called", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestRouterMethod(t=%v)", t)
}

func TestRouterGroupRouteOK(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterGroupRouteOK(t=%v)", t)
	testRouteOK(http.MethodGet, t)
	testRouteOK(http.MethodPost, t)
	testRouteOK(http.MethodPut, t)
	testRouteOK(http.MethodPatch, t)
	testRouteOK(http.MethodHead, t)
	testRouteOK(http.MethodOptions, t)
	testRouteOK(http.MethodDelete, t)
	testRouteOK(http.MethodConnect, t)
	testRouteOK(http.MethodTrace, t)
	logritelog.Default().Printf("<<< Exiting TestRouterGroupRouteOK(t=%v)", t)
}

func TestRouteNotOK(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotOK(t=%v)", t)
	testRouteNotOK(http.MethodGet, t)
	testRouteNotOK(http.MethodPost, t)
	testRouteNotOK(http.MethodPut, t)
	testRouteNotOK(http.MethodPatch, t)
	testRouteNotOK(http.MethodHead, t)
	testRouteNotOK(http.MethodOptions, t)
	testRouteNotOK(http.MethodDelete, t)
	testRouteNotOK(http.MethodConnect, t)
	testRouteNotOK(http.MethodTrace, t)
	logritelog.Default().Printf("<<< Exiting TestRouteNotOK(t=%v)", t)
}

func TestRouteNotOK2(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotOK2(t=%v)", t)
	testRouteNotOK2(http.MethodGet, t)
	testRouteNotOK2(http.MethodPost, t)
	testRouteNotOK2(http.MethodPut, t)
	testRouteNotOK2(http.MethodPatch, t)
	testRouteNotOK2(http.MethodHead, t)
	testRouteNotOK2(http.MethodOptions, t)
	testRouteNotOK2(http.MethodDelete, t)
	testRouteNotOK2(http.MethodConnect, t)
	testRouteNotOK2(http.MethodTrace, t)
	logritelog.Default().Printf("<<< Exiting TestRouteNotOK2(t=%v)", t)
}

func TestRouteRedirectTrailingSlash(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteRedirectTrailingSlash(t=%v)", t)
	router := New()
	router.RedirectFixedPath = false
	logritelog.Default().Printf("TestRouteRedirectTrailingSlash(t=%v): router.RedirectFixedPath → %v", t, router.RedirectFixedPath)
	logritelog.Default().Printf("TestRouteRedirectTrailingSlash(t=%v): router.RedirectTrailingSlash → %v", t, router.RedirectTrailingSlash)
	router.RedirectTrailingSlash = true
	router.GET("/path", func(c *Context) {})
	router.GET("/path2/", func(c *Context) {})
	router.POST("/path3", func(c *Context) {})
	router.PUT("/path4/", func(c *Context) {})

	w := PerformRequest(router, http.MethodGet, "/path/")
	assert.Equal(t, "/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2")
	assert.Equal(t, "/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3/")
	assert.Equal(t, "/path3", w.Header().Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	w = PerformRequest(router, http.MethodPut, "/path4")
	assert.Equal(t, "/path4/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path")
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2/")
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3")
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodPut, "/path4/")
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", header{Key: "X-Forwarded-Prefix", Value: "/api"})
	assert.Equal(t, "/api/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2/", header{Key: "X-Forwarded-Prefix", Value: "/api/"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "../../api#?"})
	assert.Equal(t, "/api/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "../../api"})
	assert.Equal(t, "/api/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", header{Key: "X-Forwarded-Prefix", Value: "../../api"})
	assert.Equal(t, "/api/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", header{Key: "X-Forwarded-Prefix", Value: "/../../api"})
	assert.Equal(t, "/api/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "api/../../"})
	assert.Equal(t, "//path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "api/../../../"})
	assert.Equal(t, "/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", header{Key: "X-Forwarded-Prefix", Value: "../../gin-gonic.com"})
	assert.Equal(t, "/gin-goniccom/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2", header{Key: "X-Forwarded-Prefix", Value: "/../../gin-gonic.com"})
	assert.Equal(t, "/gin-goniccom/path2/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "https://gin-gonic.com/#"})
	assert.Equal(t, "https/gin-goniccom/https/gin-goniccom/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "#api"})
	assert.Equal(t, "api/api/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "/nor-mal/#?a=1"})
	assert.Equal(t, "/nor-mal/a1/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path/", header{Key: "X-Forwarded-Prefix", Value: "/nor-mal/%2e%2e/"})
	assert.Equal(t, "/nor-mal/2e2e/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	router.RedirectTrailingSlash = false

	w = PerformRequest(router, http.MethodGet, "/path/")
	assert.Equal(t, http.StatusNotFound, w.Code)
	w = PerformRequest(router, http.MethodGet, "/path2")
	assert.Equal(t, http.StatusNotFound, w.Code)
	w = PerformRequest(router, http.MethodPost, "/path3/")
	assert.Equal(t, http.StatusNotFound, w.Code)
	w = PerformRequest(router, http.MethodPut, "/path4")
	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteRedirectTrailingSlash(t=%v)", t)
}

func TestRouteRedirectFixedPath(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteRedirectFixedPath(t=%v)", t)
	router := New()
	router.RedirectFixedPath = true
	logritelog.Default().Printf("TestRouteRedirectFixedPath(t=%v): router.RedirectFixedPath → %v", t, router.RedirectFixedPath)
	logritelog.Default().Printf("TestRouteRedirectFixedPath(t=%v): router.RedirectTrailingSlash → %v", t, router.RedirectTrailingSlash)
	router.RedirectTrailingSlash = false

	router.GET("/path", func(c *Context) {})
	router.GET("/Path2", func(c *Context) {})
	router.POST("/PATH3", func(c *Context) {})
	router.POST("/Path4/", func(c *Context) {})

	w := PerformRequest(router, http.MethodGet, "/PATH")
	assert.Equal(t, "/path", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodGet, "/path2")
	assert.Equal(t, "/Path2", w.Header().Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path3")
	assert.Equal(t, "/PATH3", w.Header().Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	w = PerformRequest(router, http.MethodPost, "/path4")
	assert.Equal(t, "/Path4/", w.Header().Get("Location"))
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteRedirectFixedPath(t=%v)", t)
}

// TestContextParamsGet tests that a parameter can be parsed from the URL.
func TestRouteParamsByName(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteParamsByName(t=%v)", t)
	name := ""
	lastName := ""
	wild := ""
	router := New()
	router.GET("/test/:name/:last_name/*wild", func(c *Context) {
		name = c.Params.ByName("name")
		lastName = c.Params.ByName("last_name")
		var ok bool
		logritelog.Default().Printf("TestRouteParamsByName(t=%v): name → %v", t, name)
		logritelog.Default().Printf("TestRouteParamsByName(t=%v): lastName → %v", t, lastName)
		wild, ok = c.Params.Get("wild")

		assert.True(t, ok)
		assert.Equal(t, name, c.Param("name"))
		assert.Equal(t, lastName, c.Param("last_name"))

		assert.Empty(t, c.Param("wtf"))
		assert.Empty(t, c.Params.ByName("wtf"))

		wtf, ok := c.Params.Get("wtf")
		assert.Empty(t, wtf)
		assert.False(t, ok)
	})

	w := PerformRequest(router, http.MethodGet, "/test/john/smith/is/super/great")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "john", name)
	assert.Equal(t, "smith", lastName)
	assert.Equal(t, "/is/super/great", wild)
	logritelog.Default().Printf("<<< Exiting TestRouteParamsByName(t=%v)", t)
}

// TestContextParamsGet tests that a parameter can be parsed from the URL even with extra slashes.
func TestRouteParamsByNameWithExtraSlash(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteParamsByNameWithExtraSlash(t=%v)", t)
	name := ""
	lastName := ""
	wild := ""
	router := New()
	router.RemoveExtraSlash = true
	logritelog.Default().Printf("TestRouteParamsByNameWithExtraSlash(t=%v): router.RemoveExtraSlash → %v", t, router.RemoveExtraSlash)
	router.GET("/test/:name/:last_name/*wild", func(c *Context) {
		name = c.Params.ByName("name")
		lastName = c.Params.ByName("last_name")
		var ok bool
		logritelog.Default().Printf("TestRouteParamsByNameWithExtraSlash(t=%v): name → %v", t, name)
		logritelog.Default().Printf("TestRouteParamsByNameWithExtraSlash(t=%v): lastName → %v", t, lastName)
		wild, ok = c.Params.Get("wild")

		assert.True(t, ok)
		assert.Equal(t, name, c.Param("name"))
		assert.Equal(t, lastName, c.Param("last_name"))

		assert.Empty(t, c.Param("wtf"))
		assert.Empty(t, c.Params.ByName("wtf"))

		wtf, ok := c.Params.Get("wtf")
		assert.Empty(t, wtf)
		assert.False(t, ok)
	})

	w := PerformRequest(router, http.MethodGet, "//test//john//smith//is//super//great")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "john", name)
	assert.Equal(t, "smith", lastName)
	assert.Equal(t, "/is/super/great", wild)
	logritelog.Default().Printf("<<< Exiting TestRouteParamsByNameWithExtraSlash(t=%v)", t)
}

// TestRouteParamsNotEmpty tests that context parameters will be set
// even if a route with params/wildcards is registered after the context
// initialisation (which happened in a previous requests).
func TestRouteParamsNotEmpty(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteParamsNotEmpty(t=%v)", t)
	name := ""
	lastName := ""
	wild := ""
	router := New()

	w := PerformRequest(router, http.MethodGet, "/test/john/smith/is/super/great")

	assert.Equal(t, http.StatusNotFound, w.Code)

	router.GET("/test/:name/:last_name/*wild", func(c *Context) {
		name = c.Params.ByName("name")
		lastName = c.Params.ByName("last_name")
		var ok bool
		logritelog.Default().Printf("TestRouteParamsNotEmpty(t=%v): name → %v", t, name)
		logritelog.Default().Printf("TestRouteParamsNotEmpty(t=%v): lastName → %v", t, lastName)
		wild, ok = c.Params.Get("wild")

		assert.True(t, ok)
		assert.Equal(t, name, c.Param("name"))
		assert.Equal(t, lastName, c.Param("last_name"))

		assert.Empty(t, c.Param("wtf"))
		assert.Empty(t, c.Params.ByName("wtf"))

		wtf, ok := c.Params.Get("wtf")
		assert.Empty(t, wtf)
		assert.False(t, ok)
	})

	w = PerformRequest(router, http.MethodGet, "/test/john/smith/is/super/great")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "john", name)
	assert.Equal(t, "smith", lastName)
	assert.Equal(t, "/is/super/great", wild)
	logritelog.Default().Printf("<<< Exiting TestRouteParamsNotEmpty(t=%v)", t)
}

// TestHandleStaticFile - ensure the static file handles properly
func TestRouteStaticFile(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteStaticFile(t=%v)", t)
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := os.CreateTemp(testRoot, "")
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in TestRouteStaticFile(t=%v): %v", t, err)
		t.Error(err)
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString("Gin Web Framework")
	require.NoError(t, err)
	f.Close()

	dir, filename := filepath.Split(f.Name())

	// SETUP gin
	router := New()
	router.Static("/using_static", dir)
	router.StaticFile("/result", f.Name())

	w := PerformRequest(router, http.MethodGet, "/using_static/"+filename)
	w2 := PerformRequest(router, http.MethodGet, "/result")

	assert.Equal(t, w, w2)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Gin Web Framework", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	w3 := PerformRequest(router, http.MethodHead, "/using_static/"+filename)
	w4 := PerformRequest(router, http.MethodHead, "/result")

	assert.Equal(t, w3, w4)
	assert.Equal(t, http.StatusOK, w3.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteStaticFile(t=%v)", t)
}

// TestHandleStaticFile - ensure the static file handles properly
func TestRouteStaticFileFS(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteStaticFileFS(t=%v)", t)
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := os.CreateTemp(testRoot, "")
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in TestRouteStaticFileFS(t=%v): %v", t, err)
		t.Error(err)
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString("Gin Web Framework")
	require.NoError(t, err)
	f.Close()

	dir, filename := filepath.Split(f.Name())
	// SETUP gin
	router := New()
	router.Static("/using_static", dir)
	router.StaticFileFS("/result_fs", filename, Dir(dir, false))

	w := PerformRequest(router, http.MethodGet, "/using_static/"+filename)
	w2 := PerformRequest(router, http.MethodGet, "/result_fs")

	assert.Equal(t, w, w2)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Gin Web Framework", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	w3 := PerformRequest(router, http.MethodHead, "/using_static/"+filename)
	w4 := PerformRequest(router, http.MethodHead, "/result_fs")

	assert.Equal(t, w3, w4)
	assert.Equal(t, http.StatusOK, w3.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteStaticFileFS(t=%v)", t)
}

// TestHandleStaticDir - ensure the root/sub dir handles properly
func TestRouteStaticListingDir(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteStaticListingDir(t=%v)", t)
	router := New()
	router.StaticFS("/", Dir("./", true))

	w := PerformRequest(router, http.MethodGet, "/")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "gin.go")
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	logritelog.Default().Printf("<<< Exiting TestRouteStaticListingDir(t=%v)", t)
}

// TestHandleHeadToDir - ensure the root/sub dir handles properly
func TestRouteStaticNoListing(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteStaticNoListing(t=%v)", t)
	router := New()
	router.Static("/", "./")

	w := PerformRequest(router, http.MethodGet, "/")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NotContains(t, w.Body.String(), "gin.go")
	logritelog.Default().Printf("<<< Exiting TestRouteStaticNoListing(t=%v)", t)
}

func TestRouterMiddlewareAndStatic(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterMiddlewareAndStatic(t=%v)", t)
	router := New()
	static := router.Group("/", func(c *Context) {
		c.Writer.Header().Add("Last-Modified", "Mon, 02 Jan 2006 15:04:05 MST")
		c.Writer.Header().Add("Expires", "Mon, 02 Jan 2006 15:04:05 MST")
		c.Writer.Header().Add("X-GIN", "Gin Framework")
	})
	static.Static("/", "./")

	w := PerformRequest(router, http.MethodGet, "/gin.go")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "package gin")
	// Content-Type='text/plain; charset=utf-8' when go version <= 1.16,
	// else, Content-Type='text/x-go; charset=utf-8'
	assert.NotEmpty(t, w.Header().Get("Content-Type"))
	assert.NotEqual(t, "Mon, 02 Jan 2006 15:04:05 MST", w.Header().Get("Last-Modified"))
	assert.Equal(t, "Mon, 02 Jan 2006 15:04:05 MST", w.Header().Get("Expires"))
	assert.Equal(t, "Gin Framework", w.Header().Get("x-GIN"))
	logritelog.Default().Printf("<<< Exiting TestRouterMiddlewareAndStatic(t=%v)", t)
}

func TestRouteNotAllowedEnabled(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotAllowedEnabled(t=%v)", t)
	router := New()
	router.HandleMethodNotAllowed = true
	logritelog.Default().Printf("TestRouteNotAllowedEnabled(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)
	router.POST("/path", func(c *Context) {})
	w := PerformRequest(router, http.MethodGet, "/path")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	router.NoMethod(func(c *Context) {
		c.String(http.StatusTeapot, "responseText")
	})
	w = PerformRequest(router, http.MethodGet, "/path")
	assert.Equal(t, "responseText", w.Body.String())
	assert.Equal(t, http.StatusTeapot, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteNotAllowedEnabled(t=%v)", t)
}

func TestRouteNotAllowedEnabled2(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotAllowedEnabled2(t=%v)", t)
	router := New()
	router.HandleMethodNotAllowed = true
	logritelog.Default().Printf("TestRouteNotAllowedEnabled2(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)
	// add one methodTree to trees
	router.addRoute(http.MethodPost, "/", HandlersChain{func(_ *Context) {}})
	router.GET("/path2", func(c *Context) {})
	w := PerformRequest(router, http.MethodPost, "/path2")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteNotAllowedEnabled2(t=%v)", t)
}

func TestRouteNotAllowedEnabled3(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotAllowedEnabled3(t=%v)", t)
	router := New()
	router.HandleMethodNotAllowed = true
	logritelog.Default().Printf("TestRouteNotAllowedEnabled3(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)
	router.GET("/path", func(c *Context) {})
	router.POST("/path", func(c *Context) {})
	w := PerformRequest(router, http.MethodPut, "/path")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	allowed := w.Header().Get("Allow")
	assert.Contains(t, allowed, http.MethodGet)
	assert.Contains(t, allowed, http.MethodPost)
	logritelog.Default().Printf("<<< Exiting TestRouteNotAllowedEnabled3(t=%v)", t)
}

func TestRouteNotAllowedDisabled(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteNotAllowedDisabled(t=%v)", t)
	router := New()
	router.HandleMethodNotAllowed = false
	logritelog.Default().Printf("TestRouteNotAllowedDisabled(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)
	router.POST("/path", func(c *Context) {})
	w := PerformRequest(router, http.MethodGet, "/path")
	assert.Equal(t, http.StatusNotFound, w.Code)

	router.NoMethod(func(c *Context) {
		c.String(http.StatusTeapot, "responseText")
	})
	w = PerformRequest(router, http.MethodGet, "/path")
	assert.Equal(t, "404 page not found", w.Body.String())
	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteNotAllowedDisabled(t=%v)", t)
}

func TestRouterNotFoundWithRemoveExtraSlash(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterNotFoundWithRemoveExtraSlash(t=%v)", t)
	router := New()
	router.RemoveExtraSlash = true
	logritelog.Default().Printf("TestRouterNotFoundWithRemoveExtraSlash(t=%v): router.RemoveExtraSlash → %v", t, router.RemoveExtraSlash)
	router.GET("/path", func(c *Context) {})
	router.GET("/", func(c *Context) {})

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/../path", http.StatusOK, ""},    // CleanPath
		{"/nope", http.StatusNotFound, ""}, // NotFound
	}
	for _, tr := range testRoutes {
		w := PerformRequest(router, http.MethodGet, tr.route)
		assert.Equal(t, tr.code, w.Code)
		if w.Code != http.StatusNotFound {
			assert.Equal(t, tr.location, w.Header().Get("Location"))
		}
	}
	logritelog.Default().Printf("<<< Exiting TestRouterNotFoundWithRemoveExtraSlash(t=%v)", t)
}

func TestRouterNotFound(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterNotFound(t=%v)", t)
	router := New()
	router.RedirectFixedPath = true
	logritelog.Default().Printf("TestRouterNotFound(t=%v): router.RedirectFixedPath → %v", t, router.RedirectFixedPath)
	router.GET("/path", func(c *Context) {})
	router.GET("/dir/", func(c *Context) {})
	router.GET("/", func(c *Context) {})

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // Without CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		w := PerformRequest(router, http.MethodGet, tr.route)
		assert.Equal(t, tr.code, w.Code)
		if w.Code != http.StatusNotFound {
			assert.Equal(t, tr.location, w.Header().Get("Location"))
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NoRoute(func(c *Context) {
		c.AbortWithStatus(http.StatusNotFound)
		notFound = true
		logritelog.Default().Printf("TestRouterNotFound(t=%v): notFound → %v", t, notFound)
	})
	w := PerformRequest(router, http.MethodGet, "/nope")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.True(t, notFound)

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", func(c *Context) {})
	w = PerformRequest(router, http.MethodPatch, "/path/")
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "map[Location:[/path]]", fmt.Sprint(w.Header()))

	// Test special case where no node for the prefix "/" exists
	router = New()
	router.GET("/a", func(c *Context) {})
	w = PerformRequest(router, http.MethodGet, "/")
	assert.Equal(t, http.StatusNotFound, w.Code)

	logritelog.Default().Printf("TestRouterNotFound(t=%v): router → %v", t, router)
	// Reproduction test for the bug of issue #2843
	router = New()
	router.NoRoute(func(c *Context) {
		if c.Request.RequestURI == "/login" {
			c.String(http.StatusOK, "login")
		}
	})
	router.GET("/logout", func(c *Context) {
		c.String(http.StatusOK, "logout")
	})
	w = PerformRequest(router, http.MethodGet, "/login")
	assert.Equal(t, "login", w.Body.String())
	w = PerformRequest(router, http.MethodGet, "/logout")
	assert.Equal(t, "logout", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestRouterNotFound(t=%v)", t)
}

func TestRouterStaticFSNotFound(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterStaticFSNotFound(t=%v)", t)
	router := New()
	router.StaticFS("/", http.FileSystem(http.Dir("/thisreallydoesntexist/")))
	router.NoRoute(func(c *Context) {
		c.String(http.StatusNotFound, "non existent")
	})

	w := PerformRequest(router, http.MethodGet, "/nonexistent")
	assert.Equal(t, "non existent", w.Body.String())

	w = PerformRequest(router, http.MethodHead, "/nonexistent")
	assert.Equal(t, "non existent", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestRouterStaticFSNotFound(t=%v)", t)
}

func TestRouterStaticFSFileNotFound(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouterStaticFSFileNotFound(t=%v)", t)
	router := New()

	router.StaticFS("/", http.FileSystem(http.Dir(".")))

	assert.NotPanics(t, func() {
		PerformRequest(router, http.MethodGet, "/nonexistent")
	})
	logritelog.Default().Printf("<<< Exiting TestRouterStaticFSFileNotFound(t=%v)", t)
}

// Reproduction test for the bug of issue #1805
func TestMiddlewareCalledOnceByRouterStaticFSNotFound(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestMiddlewareCalledOnceByRouterStaticFSNotFound(t=%v)", t)
	router := New()

	// Middleware must be called just only once by per request.
	middlewareCalledNum := 0
	router.Use(func(c *Context) {
		middlewareCalledNum++
	})

	router.StaticFS("/", http.FileSystem(http.Dir("/thisreallydoesntexist/")))

	// First access
	PerformRequest(router, http.MethodGet, "/nonexistent")
	assert.Equal(t, 1, middlewareCalledNum)

	// Second access
	PerformRequest(router, http.MethodHead, "/nonexistent")
	assert.Equal(t, 2, middlewareCalledNum)
	logritelog.Default().Printf("<<< Exiting TestMiddlewareCalledOnceByRouterStaticFSNotFound(t=%v)", t)
}

func TestRouteRawPath(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteRawPath(t=%v)", t)
	route := New()
	route.UseRawPath = true
	logritelog.Default().Printf("TestRouteRawPath(t=%v): route.UseRawPath → %v", t, route.UseRawPath)

	route.POST("/project/:name/build/:num", func(c *Context) {
		name := c.Params.ByName("name")
		num := c.Params.ByName("num")

		assert.Equal(t, name, c.Param("name"))
		assert.Equal(t, num, c.Param("num"))

		assert.Equal(t, "Some/Other/Project", name)
		assert.Equal(t, "222", num)
	})

	w := PerformRequest(route, http.MethodPost, "/project/Some%2FOther%2FProject/build/222")
	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteRawPath(t=%v)", t)
}

func TestRouteRawPathNoUnescape(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteRawPathNoUnescape(t=%v)", t)
	route := New()
	route.UseRawPath = true
	logritelog.Default().Printf("TestRouteRawPathNoUnescape(t=%v): route.UseRawPath → %v", t, route.UseRawPath)
	logritelog.Default().Printf("TestRouteRawPathNoUnescape(t=%v): route.UnescapePathValues → %v", t, route.UnescapePathValues)
	route.UnescapePathValues = false

	route.POST("/project/:name/build/:num", func(c *Context) {
		name := c.Params.ByName("name")
		num := c.Params.ByName("num")

		assert.Equal(t, name, c.Param("name"))
		assert.Equal(t, num, c.Param("num"))

		assert.Equal(t, "Some%2FOther%2FProject", name)
		assert.Equal(t, "333", num)
	})

	w := PerformRequest(route, http.MethodPost, "/project/Some%2FOther%2FProject/build/333")
	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteRawPathNoUnescape(t=%v)", t)
}

func TestRouteServeErrorWithWriteHeader(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteServeErrorWithWriteHeader(t=%v)", t)
	route := New()
	route.Use(func(c *Context) {
		c.Status(http.StatusMisdirectedRequest)
		c.Next()
	})

	w := PerformRequest(route, http.MethodGet, "/NotFound")
	assert.Equal(t, http.StatusMisdirectedRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	logritelog.Default().Printf("<<< Exiting TestRouteServeErrorWithWriteHeader(t=%v)", t)
}

func TestRouteContextHoldsFullPath(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRouteContextHoldsFullPath(t=%v)", t)
	router := New()

	// Test routes
	routes := []string{
		"/simple",
		"/project/:name",
		"/",
		"/news/home",
		"/news",
		"/simple-two/one",
		"/simple-two/one-two",
		"/project/:name/build/*params",
		"/project/:name/bui",
		"/user/:id/status",
		"/user/:id",
		"/user/:id/profile",
	}

	for _, route := range routes {
		actualRoute := route
		router.GET(route, func(c *Context) {
			// For each defined route context should contain its full path
			assert.Equal(t, actualRoute, c.FullPath())
			c.AbortWithStatus(http.StatusOK)
		})
	}

	for _, route := range routes {
		w := PerformRequest(router, http.MethodGet, route)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test not found
	router.Use(func(c *Context) {
		// For not found routes full path is empty
		assert.Empty(t, c.FullPath())
	})

	w := PerformRequest(router, http.MethodGet, "/not-found")
	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("<<< Exiting TestRouteContextHoldsFullPath(t=%v)", t)
}

func TestEngineHandleMethodNotAllowedCornerCase(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestEngineHandleMethodNotAllowedCornerCase(t=%v)", t)
	r := New()
	r.HandleMethodNotAllowed = true
	logritelog.Default().Printf("TestEngineHandleMethodNotAllowedCornerCase(t=%v): r.HandleMethodNotAllowed → %v", t, r.HandleMethodNotAllowed)

	base := r.Group("base")
	base.GET("/metrics", handlerTest1)

	v1 := base.Group("v1")

	v1.GET("/:id/devices", handlerTest1)
	v1.GET("/user/:id/groups", handlerTest1)

	v1.GET("/orgs/:id", handlerTest1)
	v1.DELETE("/orgs/:id", handlerTest1)

	w := PerformRequest(r, http.MethodGet, "/base/v1/user/groups")
	assert.Equal(t, http.StatusNotFound, w.Code)
	logritelog.Default().Printf("<<< Exiting TestEngineHandleMethodNotAllowedCornerCase(t=%v)", t)
}
