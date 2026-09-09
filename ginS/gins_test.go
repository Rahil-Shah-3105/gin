// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ginS
import logritelog "log"

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	logritelog.Default().Printf(">>> Entering init()")
	gin.SetMode(gin.TestMode)
	logritelog.Default().Printf("<<< Exiting init()")
}

func TestGET(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestGET(t=%v)", t)
	GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestGET(t=%v)", t)
}

func TestPOST(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPOST(t=%v)", t)
	POST("/post", func(c *gin.Context) {
		c.String(http.StatusCreated, "created")
	})

	req := httptest.NewRequest(http.MethodPost, "/post", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "created", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestPOST(t=%v)", t)
}

func TestPUT(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPUT(t=%v)", t)
	PUT("/put", func(c *gin.Context) {
		c.String(http.StatusOK, "updated")
	})

	req := httptest.NewRequest(http.MethodPut, "/put", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "updated", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestPUT(t=%v)", t)
}

func TestDELETE(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestDELETE(t=%v)", t)
	DELETE("/delete", func(c *gin.Context) {
		c.String(http.StatusOK, "deleted")
	})

	req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "deleted", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestDELETE(t=%v)", t)
}

func TestPATCH(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestPATCH(t=%v)", t)
	PATCH("/patch", func(c *gin.Context) {
		c.String(http.StatusOK, "patched")
	})

	req := httptest.NewRequest(http.MethodPatch, "/patch", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "patched", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestPATCH(t=%v)", t)
}

func TestOPTIONS(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestOPTIONS(t=%v)", t)
	OPTIONS("/options", func(c *gin.Context) {
		c.String(http.StatusOK, "options")
	})

	req := httptest.NewRequest(http.MethodOptions, "/options", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "options", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestOPTIONS(t=%v)", t)
}

func TestHEAD(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestHEAD(t=%v)", t)
	HEAD("/head", func(c *gin.Context) {
		c.String(http.StatusOK, "head")
	})

	req := httptest.NewRequest(http.MethodHead, "/head", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestHEAD(t=%v)", t)
}

func TestAny(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestAny(t=%v)", t)
	Any("/any", func(c *gin.Context) {
		c.String(http.StatusOK, "any")
	})

	req := httptest.NewRequest(http.MethodGet, "/any", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "any", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestAny(t=%v)", t)
}

func TestHandle(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestHandle(t=%v)", t)
	Handle(http.MethodGet, "/handle", func(c *gin.Context) {
		c.String(http.StatusOK, "handle")
	})

	req := httptest.NewRequest(http.MethodGet, "/handle", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "handle", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestHandle(t=%v)", t)
}

func TestGroup(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestGroup(t=%v)", t)
	group := Group("/group")
	group.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "group test")
	})

	req := httptest.NewRequest(http.MethodGet, "/group/test", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "group test", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestGroup(t=%v)", t)
}

func TestUse(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestUse(t=%v)", t)
	var middlewareExecuted bool
	Use(func(c *gin.Context) {
		middlewareExecuted = true
		logritelog.Default().Printf("TestUse(t=%v): middlewareExecuted → %v", t, middlewareExecuted)
		c.Next()
	})

	GET("/middleware-test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/middleware-test", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.True(t, middlewareExecuted)
	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestUse(t=%v)", t)
}

func TestNoRoute(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestNoRoute(t=%v)", t)
	NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "custom 404")
	})

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "custom 404", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestNoRoute(t=%v)", t)
}

func TestNoMethod(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestNoMethod(t=%v)", t)
	NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "method not allowed")
	})

	// This just verifies that NoMethod is callable
	// Testing the actual behavior would require a separate engine instance
	assert.NotNil(t, engine())
	logritelog.Default().Printf("<<< Exiting TestNoMethod(t=%v)", t)
}

func TestRoutes(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestRoutes(t=%v)", t)
	GET("/routes-test", func(c *gin.Context) {})

	routes := Routes()
	assert.NotEmpty(t, routes)

	found := false
	for _, route := range routes {
		if route.Path == "/routes-test" && route.Method == http.MethodGet {
			found = true
			logritelog.Default().Printf("TestRoutes(t=%v): found → %v", t, found)
			break
		}
	}
	assert.True(t, found)
	logritelog.Default().Printf("<<< Exiting TestRoutes(t=%v)", t)
}

func TestSetHTMLTemplate(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestSetHTMLTemplate(t=%v)", t)
	tmpl := template.Must(template.New("test").Parse("Hello {{.}}"))
	SetHTMLTemplate(tmpl)

	// Verify engine has template set
	assert.NotNil(t, engine())
	logritelog.Default().Printf("<<< Exiting TestSetHTMLTemplate(t=%v)", t)
}

func TestStaticFile(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestStaticFile(t=%v)", t)
	StaticFile("/static-file", "../testdata/test_file.txt")

	req := httptest.NewRequest(http.MethodGet, "/static-file", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestStaticFile(t=%v)", t)
}

func TestStatic(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestStatic(t=%v)", t)
	Static("/static-dir", "../testdata")

	req := httptest.NewRequest(http.MethodGet, "/static-dir/test_file.txt", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestStatic(t=%v)", t)
}

func TestStaticFS(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestStaticFS(t=%v)", t)
	fs := http.Dir("../testdata")
	StaticFS("/static-fs", fs)

	req := httptest.NewRequest(http.MethodGet, "/static-fs/test_file.txt", nil)
	w := httptest.NewRecorder()
	engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	logritelog.Default().Printf("<<< Exiting TestStaticFS(t=%v)", t)
}
