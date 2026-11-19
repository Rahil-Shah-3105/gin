// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDebugging(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestIsDebugging(t=%v)", t)
	SetMode(DebugMode)
	assert.True(t, IsDebugging())
	SetMode(ReleaseMode)
	assert.False(t, IsDebugging())
	SetMode(TestMode)
	assert.False(t, IsDebugging())
	logritelog.Default().Printf("[WARN] <<< Exiting TestIsDebugging(t=%v)", t)
}

func TestDebugPrint(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrint(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		SetMode(ReleaseMode)
		debugPrint("DEBUG this!")
		SetMode(TestMode)
		debugPrint("DEBUG this!")
		SetMode(DebugMode)
		debugPrint("these are %d %s", 2, "error messages")
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] these are 2 error messages\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrint(t=%v)", t)
}

func TestDebugPrintFunc(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintFunc(t=%v)", t)
	DebugPrintFunc = func(format string, values ...any) {
		fmt.Fprintf(DefaultWriter, "[GIN-debug] "+format, values...)
	}
		logritelog.Default().Printf("[DEBUG] TestDebugPrintFunc(t=%v): DebugPrintFunc → %v", t, DebugPrintFunc)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrint("debug print func test: %d", 123)
		SetMode(TestMode)
	})
	assert.Regexp(t, `^\[GIN-debug\] debug print func test: 123`, re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintFunc(t=%v)", t)
}

func TestDebugPrintError(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintError(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintError(nil)
		debugPrintError(errors.New("this is an error"))
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] [ERROR] this is an error\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintError(t=%v)", t)
}

func TestDebugPrintRoutes(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintRoutes(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintRoute(http.MethodGet, "/path/to/route/:param", HandlersChain{func(c *Context) {}, handlerNameTest})
		SetMode(TestMode)
	})
	assert.Regexp(t, `^\[GIN-debug\] GET    /path/to/route/:param     --> (.*/vendor/)?github.com/gin-gonic/gin.handlerNameTest \(2 handlers\)\n$`, re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintRoutes(t=%v)", t)
}

func TestDebugPrintRouteFunc(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintRouteFunc(t=%v)", t)
	DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		fmt.Fprintf(DefaultWriter, "[GIN-debug] %-6s %-40s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
		logritelog.Default().Printf("[DEBUG] TestDebugPrintRouteFunc(t=%v): DebugPrintRouteFunc → %v", t, DebugPrintRouteFunc)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintRoute(http.MethodGet, "/path/to/route/:param1/:param2", HandlersChain{func(c *Context) {}, handlerNameTest})
		SetMode(TestMode)
	})
	assert.Regexp(t, `^\[GIN-debug\] GET    /path/to/route/:param1/:param2           --> (.*/vendor/)?github.com/gin-gonic/gin.handlerNameTest \(2 handlers\)\n$`, re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintRouteFunc(t=%v)", t)
}

func TestDebugPrintLoadTemplate(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintLoadTemplate(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		templ := template.Must(template.New("").Delims("{[{", "}]}").ParseGlob("./testdata/template/hello.tmpl"))
		debugPrintLoadTemplate(templ)
		SetMode(TestMode)
	})
	assert.Regexp(t, `^\[GIN-debug\] Loaded HTML Templates \(2\): \n(\t- \n|\t- hello\.tmpl\n){2}\n`, re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintLoadTemplate(t=%v)", t)
}

func TestDebugPrintWARNINGSetHTMLTemplate(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintWARNINGSetHTMLTemplate(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintWARNINGSetHTMLTemplate()
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] [WARNING] Since SetHTMLTemplate() is NOT thread-safe. It should only be called\nat initialization. ie. before any route is registered or the router is listening in a socket:\n\n\trouter := gin.Default()\n\trouter.SetHTMLTemplate(template) // << good place\n\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintWARNINGSetHTMLTemplate(t=%v)", t)
}

func TestDebugPrintWARNINGDefault(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintWARNINGDefault(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintWARNINGDefault()
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintWARNINGDefault(t=%v)", t)
}

func TestDebugPrintWARNINGDefaultWithUnsupportedVersion(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintWARNINGDefaultWithUnsupportedVersion(t=%v)", t)
	runtimeVersion = "go1.23.12"
	logritelog.Default().Printf("[DEBUG] TestDebugPrintWARNINGDefaultWithUnsupportedVersion(t=%v): runtimeVersion → %v", t, runtimeVersion)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintWARNINGDefault()
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] [WARNING] Now Gin requires Go 1.25+.\n\n[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintWARNINGDefaultWithUnsupportedVersion(t=%v)", t)
}

func TestDebugPrintWARNINGNew(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDebugPrintWARNINGNew(t=%v)", t)
	re := captureOutput(t, func() {
		SetMode(DebugMode)
		debugPrintWARNINGNew()
		SetMode(TestMode)
	})
	assert.Equal(t, "[GIN-debug] [WARNING] Running in \"debug\" mode. Switch to \"release\" mode in production.\n - using env:\texport GIN_MODE=release\n - using code:\tgin.SetMode(gin.ReleaseMode)\n\n", re)
	logritelog.Default().Printf("[WARN] <<< Exiting TestDebugPrintWARNINGNew(t=%v)", t)
}

func captureOutput(t *testing.T, f func()) string {
	logritelog.Default().Printf("[ERROR] >>> Entering captureOutput(t=%v,f=%v)", t, f)
	reader, writer, err := os.Pipe()
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in captureOutput(t=%v,f=%v): %v", t, f, err)
		panic(err)
	}
	defaultWriter := DefaultWriter
	logritelog.Default().Printf("[WARN] <<< Exiting captureOutput(t=%v,f=%v)", t, f)
	defaultErrorWriter := DefaultErrorWriter
	defer func() {
		DefaultWriter = defaultWriter
		DefaultErrorWriter = defaultErrorWriter
		log.SetOutput(os.Stderr)
	}()
	DefaultWriter = writer
	DefaultErrorWriter = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf strings.Builder
		wg.Done()
		_, err := io.Copy(&buf, reader)
		assert.NoError(t, err)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func TestGetMinVer(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestGetMinVer(t=%v)", t)
	var m uint64
	var e error
	_, e = getMinVer("go1")
	require.Error(t, e)
	m, e = getMinVer("go1.1")
	assert.Equal(t, uint64(1), m)
	require.NoError(t, e)
	m, e = getMinVer("go1.1.1")
	require.NoError(t, e)
	assert.Equal(t, uint64(1), m)
	_, e = getMinVer("go1.1.1.1")
	require.Error(t, e)
	logritelog.Default().Printf("[WARN] <<< Exiting TestGetMinVer(t=%v)", t)
}
