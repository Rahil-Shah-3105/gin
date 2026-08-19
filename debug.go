// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"fmt"
	"html/template"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

const ginSupportMinGoVer = 25

var runtimeVersion = runtime.Version()

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	logritelog.Default().Printf(">>> Entering IsDebugging()")
	return atomic.LoadInt32(&ginMode) == debugCode
}

// DebugPrintRouteFunc indicates debug log output format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

// DebugPrintFunc indicates debug log output format.
var DebugPrintFunc func(format string, values ...any)

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())
		if DebugPrintRouteFunc == nil {
			debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrintLoadTemplate(tmpl *template.Template) {
	if IsDebugging() {
		var buf strings.Builder
		for _, tmpl := range tmpl.Templates() {
			buf.WriteString("\t- ")
			buf.WriteString(tmpl.Name())
			buf.WriteString("\n")
		}
		debugPrint("Loaded HTML Templates (%d): \n%s\n", len(tmpl.Templates()), buf.String())
	}
}

func debugPrint(format string, values ...any) {
	logritelog.Default().Printf(">>> Entering debugPrint(format=%v,values=%v)", format, values)
	if !IsDebugging() {
		return
	}

	if DebugPrintFunc != nil {
		DebugPrintFunc(format, values...)
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(DefaultWriter, "[GIN-debug] "+format, values...)
	logritelog.Default().Printf("<<< Exiting debugPrint(format=%v,values=%v)", format, values)
}

func getMinVer(v string) (uint64, error) {
	logritelog.Default().Printf(">>> Entering getMinVer(v=%v)", v)
	first := strings.IndexByte(v, '.')
	last := strings.LastIndexByte(v, '.')
	if first == last {
	logritelog.Default().Printf("<<< Exiting getMinVer(v=%v)", v)
		return strconv.ParseUint(v[first+1:], 10, 64)
	}
	return strconv.ParseUint(v[first+1:last], 10, 64)
}

func debugPrintWARNINGDefault() {
	logritelog.Default().Printf(">>> Entering debugPrintWARNINGDefault()")
	if v, e := getMinVer(runtimeVersion); e == nil && v < ginSupportMinGoVer {
	logritelog.Default().Printf("<<< Exiting debugPrintWARNINGDefault()")
		debugPrint(`[WARNING] Now Gin requires Go 1.25+.

`)
	}
	debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

`)
}

func debugPrintWARNINGNew() {
	logritelog.Default().Printf(">>> Entering debugPrintWARNINGNew()")
	logritelog.Default().Printf("<<< Exiting debugPrintWARNINGNew()")
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

`)
}

func debugPrintWARNINGSetHTMLTemplate() {
	logritelog.Default().Printf(">>> Entering debugPrintWARNINGSetHTMLTemplate()")
	logritelog.Default().Printf("<<< Exiting debugPrintWARNINGSetHTMLTemplate()")
	debugPrint(`[WARNING] Since SetHTMLTemplate() is NOT thread-safe. It should only be called
at initialization. ie. before any route is registered or the router is listening in a socket:

	router := gin.Default()
	router.SetHTMLTemplate(template) // << good place

`)
}

func debugPrintError(err error) {
	logritelog.Default().Printf(">>> Entering debugPrintError(err=%v)", err)
	if err != nil && IsDebugging() {
		fmt.Fprintf(DefaultErrorWriter, "[GIN-debug] [ERROR] %v\n", err)
	}
	logritelog.Default().Printf("<<< Exiting debugPrintError(err=%v)", err)
}
