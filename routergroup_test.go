// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var MaxHandlers = 32

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	SetMode(TestMode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

func TestRouterGroupBasic(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupBasic(t=%v)", t)
	router := New()
	group := router.Group("/hola", func(c *Context) {})
	group.Use(func(c *Context) {})

	assert.Len(t, group.Handlers, 2)
	assert.Equal(t, "/hola", group.BasePath())
	assert.Equal(t, router, group.engine)

	group2 := group.Group("manu")
	group2.Use(func(c *Context) {}, func(c *Context) {})

	assert.Len(t, group2.Handlers, 4)
	assert.Equal(t, "/hola/manu", group2.BasePath())
	assert.Equal(t, router, group2.engine)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupBasic(t=%v)", t)
}

func TestRouterGroupBasicHandle(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupBasicHandle(t=%v)", t)
	performRequestInGroup(t, http.MethodGet)
	performRequestInGroup(t, http.MethodPost)
	performRequestInGroup(t, http.MethodPut)
	performRequestInGroup(t, http.MethodPatch)
	performRequestInGroup(t, http.MethodDelete)
	performRequestInGroup(t, http.MethodHead)
	performRequestInGroup(t, http.MethodOptions)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupBasicHandle(t=%v)", t)
}

func performRequestInGroup(t *testing.T, method string) {
	logritelog.Default().Printf("[ERROR] >>> Entering performRequestInGroup(t=%v,method=%v)", t, method)
	router := New()
	v1 := router.Group("v1", func(c *Context) {})
	assert.Equal(t, "/v1", v1.BasePath())

	login := v1.Group("/login/", func(c *Context) {}, func(c *Context) {})
	assert.Equal(t, "/v1/login/", login.BasePath())

	handler := func(c *Context) {
		c.String(http.StatusBadRequest, "the method was %s and index %d", c.Request.Method, c.index)
	}

	switch method {
	case http.MethodGet:
		v1.GET("/test", handler)
		login.GET("/test", handler)
	case http.MethodPost:
		v1.POST("/test", handler)
		login.POST("/test", handler)
	case http.MethodPut:
		v1.PUT("/test", handler)
		login.PUT("/test", handler)
	case http.MethodPatch:
		v1.PATCH("/test", handler)
		login.PATCH("/test", handler)
	case http.MethodDelete:
		v1.DELETE("/test", handler)
		login.DELETE("/test", handler)
	case http.MethodHead:
		v1.HEAD("/test", handler)
		login.HEAD("/test", handler)
	case http.MethodOptions:
		v1.OPTIONS("/test", handler)
		login.OPTIONS("/test", handler)
	default:
		panic("unknown method")
	}

	w := PerformRequest(router, method, "/v1/login/test")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "the method was "+method+" and index 3", w.Body.String())

	w = PerformRequest(router, method, "/v1/test")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "the method was "+method+" and index 1", w.Body.String())
	logritelog.Default().Printf("[WARN] <<< Exiting performRequestInGroup(t=%v,method=%v)", t, method)
}

func TestRouterGroupInvalidStatic(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupInvalidStatic(t=%v)", t)
	router := New()
	assert.Panics(t, func() {
		router.Static("/path/:param", "/")
	})

	assert.Panics(t, func() {
		router.Static("/path/*param", "/")
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupInvalidStatic(t=%v)", t)
}

func TestRouterGroupInvalidStaticFile(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupInvalidStaticFile(t=%v)", t)
	router := New()
	assert.Panics(t, func() {
		router.StaticFile("/path/:param", "favicon.ico")
	})

	assert.Panics(t, func() {
		router.StaticFile("/path/*param", "favicon.ico")
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupInvalidStaticFile(t=%v)", t)
}

func TestRouterGroupInvalidStaticFileFS(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupInvalidStaticFileFS(t=%v)", t)
	router := New()
	assert.Panics(t, func() {
		router.StaticFileFS("/path/:param", "favicon.ico", Dir(".", false))
	})

	assert.Panics(t, func() {
		router.StaticFileFS("/path/*param", "favicon.ico", Dir(".", false))
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupInvalidStaticFileFS(t=%v)", t)
}

func TestRouterGroupTooManyHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupTooManyHandlers(t=%v)", t)
	const (
		panicValue = "too many handlers"
		maximumCnt = abortIndex
	)
	router := New()
	handlers1 := make([]HandlerFunc, maximumCnt-1)
	router.Use(handlers1...)

	handlers2 := make([]HandlerFunc, maximumCnt+1)
	assert.PanicsWithValue(t, panicValue, func() {
		router.Use(handlers2...)
	})
	assert.PanicsWithValue(t, panicValue, func() {
		router.GET("/", handlers2...)
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupTooManyHandlers(t=%v)", t)
}

func TestRouterGroupBadMethod(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupBadMethod(t=%v)", t)
	router := New()
	assert.Panics(t, func() {
		router.Handle(http.MethodGet, "/")
	})
	assert.Panics(t, func() {
		router.Handle(" GET", "/")
	})
	assert.Panics(t, func() {
		router.Handle("GET ", "/")
	})
	assert.Panics(t, func() {
		router.Handle("", "/")
	})
	assert.Panics(t, func() {
		router.Handle("PO ST", "/")
	})
	assert.Panics(t, func() {
		router.Handle("1GET", "/")
	})
	assert.Panics(t, func() {
		router.Handle("PATCh", "/")
	})
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupBadMethod(t=%v)", t)
}

func TestRouterGroupPipeline(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupPipeline(t=%v)", t)
	router := New()
	testRoutesInterface(t, router)

	v1 := router.Group("/v1")
	testRoutesInterface(t, v1)
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupPipeline(t=%v)", t)
}

func testRoutesInterface(t *testing.T, r IRoutes) {
	logritelog.Default().Printf("[ERROR] >>> Entering testRoutesInterface(t=%v,r=%v)", t, r)
	handler := func(c *Context) {}
	assert.Equal(t, r, r.Use(handler))

	assert.Equal(t, r, r.Handle(http.MethodGet, "/handler", handler))
	assert.Equal(t, r, r.Any("/any", handler))
	assert.Equal(t, r, r.GET("/", handler))
	assert.Equal(t, r, r.POST("/", handler))
	assert.Equal(t, r, r.DELETE("/", handler))
	assert.Equal(t, r, r.PATCH("/", handler))
	assert.Equal(t, r, r.PUT("/", handler))
	assert.Equal(t, r, r.OPTIONS("/", handler))
	assert.Equal(t, r, r.HEAD("/", handler))
	assert.Equal(t, r, r.Match([]string{http.MethodPut, http.MethodPatch}, "/match", handler))

	assert.Equal(t, r, r.StaticFile("/file", "."))
	assert.Equal(t, r, r.StaticFileFS("/static2", ".", Dir(".", false)))
	assert.Equal(t, r, r.Static("/static", "."))
	assert.Equal(t, r, r.StaticFS("/static2", Dir(".", false)))
	logritelog.Default().Printf("[WARN] <<< Exiting testRoutesInterface(t=%v,r=%v)", t, r)
}

func TestRouterGroupCombineHandlersTooManyHandlers(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupCombineHandlersTooManyHandlers(t=%v)", t)
	group := &RouterGroup{
		Handlers: make(HandlersChain, MaxHandlers), // Assume group already has MaxHandlers middleware
	}
	tooManyHandlers := make(HandlersChain, MaxHandlers) // Add MaxHandlers more, total 2 * MaxHandlers

	// This should trigger panic
	assert.Panics(t, func() {
		group.combineHandlers(tooManyHandlers)
	}, "should panic due to too many handlers")
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupCombineHandlersTooManyHandlers(t=%v)", t)
}

func TestRouterGroupCombineHandlersEmptySliceNotNil(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestRouterGroupCombineHandlersEmptySliceNotNil(t=%v)", t)
	group := &RouterGroup{
		Handlers: HandlersChain{},
	}

	result := group.combineHandlers(HandlersChain{})
	assert.NotNil(t, result, "result should not be nil even with empty handlers")
	assert.Empty(t, result, "empty handlers should return empty chain")
	logritelog.Default().Printf("[WARN] <<< Exiting TestRouterGroupCombineHandlersEmptySliceNotNil(t=%v)", t)
}
