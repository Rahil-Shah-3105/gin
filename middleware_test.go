// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-contrib/sse"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareGeneralCase(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareGeneralCase(t=%v)", t)
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
	})
	router.GET("/", func(c *Context) {
		signature += "D"
	})
	router.NoRoute(func(c *Context) {
		signature += " X "
	})
	router.NoMethod(func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ACDB", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareGeneralCase(t=%v)", t)
}

func TestMiddlewareNoRoute(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareNoRoute(t=%v)", t)
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		c.Next()
		c.Next()
		c.Next()
		signature += "D"
	})
	router.NoRoute(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NoMethod(func(c *Context) {
		signature += " X "
	})
	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "ACEGHFDB", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareNoRoute(t=%v)", t)
}

func TestMiddlewareNoMethodEnabled(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareNoMethodEnabled(t=%v)", t)
	signature := ""
	router := New()
	router.HandleMethodNotAllowed = true
	logritelog.Default().Printf("[DEBUG] TestMiddlewareNoMethodEnabled(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		signature += "D"
	})
	router.NoMethod(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NoRoute(func(c *Context) {
		signature += " X "
	})
	router.POST("/", func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "ACEGHFDB", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareNoMethodEnabled(t=%v)", t)
}

func TestMiddlewareNoMethodDisabled(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareNoMethodDisabled(t=%v)", t)
	signature := ""
	router := New()

	// NoMethod disabled
	router.HandleMethodNotAllowed = false
	logritelog.Default().Printf("[DEBUG] TestMiddlewareNoMethodDisabled(t=%v): router.HandleMethodNotAllowed → %v", t, router.HandleMethodNotAllowed)

	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Next()
		signature += "D"
	})
	router.NoMethod(func(c *Context) {
		signature += "E"
		c.Next()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Next()
		signature += "H"
	})
	router.NoRoute(func(c *Context) {
		signature += " X "
	})
	router.POST("/", func(c *Context) {
		signature += " XX "
	})

	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "AC X DB", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareNoMethodDisabled(t=%v)", t)
}

func TestMiddlewareAbort(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareAbort(t=%v)", t)
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.AbortWithStatus(http.StatusUnauthorized)
		c.Next()
		signature += "D"
	})
	router.GET("/", func(c *Context) {
		signature += " X "
		c.Next()
		signature += " XX "
	})

	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "ACD", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareAbort(t=%v)", t)
}

func TestMiddlewareAbortHandlersChainAndNext(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareAbortHandlersChainAndNext(t=%v)", t)
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Next()
		c.AbortWithStatus(http.StatusGone)
		signature += "B"
	})
	router.GET("/", func(c *Context) {
		signature += "C"
		c.Next()
	})
	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusGone, w.Code)
	assert.Equal(t, "ACB", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareAbortHandlersChainAndNext(t=%v)", t)
}

// TestMiddlewareFailHandlersChain - ensure that Fail interrupt used middleware in fifo order as
// as well as Abort
func TestMiddlewareFailHandlersChain(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareFailHandlersChain(t=%v)", t)
	// SETUP
	signature := ""
	router := New()
	router.Use(func(context *Context) {
		signature += "A"
		context.AbortWithError(http.StatusInternalServerError, errors.New("foo")) //nolint: errcheck
	})
	router.Use(func(context *Context) {
		signature += "B"
		context.Next()
		signature += "C"
	})
	// RUN
	w := PerformRequest(router, http.MethodGet, "/")

	// TEST
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "A", signature)
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareFailHandlersChain(t=%v)", t)
}

func TestMiddlewareWrite(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestMiddlewareWrite(t=%v)", t)
	router := New()
	router.Use(func(c *Context) {
		c.String(http.StatusBadRequest, "hola\n")
	})
	router.Use(func(c *Context) {
		c.XML(http.StatusBadRequest, H{"foo": "bar"})
	})
	router.Use(func(c *Context) {
		c.JSON(http.StatusBadRequest, H{"foo": "bar"})
	})
	router.GET("/", func(c *Context) {
		c.JSON(http.StatusBadRequest, H{"foo": "bar"})
	}, func(c *Context) {
		c.Render(http.StatusBadRequest, sse.Event{
			Event: "test",
			Data:  "message",
		})
	})

	w := PerformRequest(router, http.MethodGet, "/")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, strings.ReplaceAll("hola\n<map><foo>bar</foo></map>{\"foo\":\"bar\"}{\"foo\":\"bar\"}event:test\ndata:message\n\n", " ", ""), strings.ReplaceAll(w.Body.String(), " ", ""))
	logritelog.Default().Printf("[WARN] <<< Exiting TestMiddlewareWrite(t=%v)", t)
}
