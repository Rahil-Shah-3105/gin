// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	logritelog.Default().Printf(">>> Entering init()")
	SetMode(TestMode)
	logritelog.Default().Printf("<<< Exiting init()")
}

func TestLogger(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLogger(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithWriter(buffer))
	router.GET("/example", func(c *Context) {})
	router.POST("/example", func(c *Context) {})
	router.PUT("/example", func(c *Context) {})
	router.DELETE("/example", func(c *Context) {})
	router.PATCH("/example", func(c *Context) {})
	router.HEAD("/example", func(c *Context) {})
	router.OPTIONS("/example", func(c *Context) {})

	PerformRequest(router, http.MethodGet, "/example?a=100")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	// I wrote these first (extending the above) but then realized they are more
	// like integration tests because they test the whole logging process rather
	// than individual functions. I'm not sure where these should go.
	buffer.Reset()
	PerformRequest(router, http.MethodPost, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodPost)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodPut, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodPut)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodDelete, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodDelete)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "PATCH", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PATCH")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodGet, "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/notfound")
	logritelog.Default().Printf("<<< Exiting TestLogger(t=%v)", t)
}

func TestLoggerWithConfig(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithConfig(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithConfig(LoggerConfig{Output: buffer}))
	router.GET("/example", func(c *Context) {})
	router.POST("/example", func(c *Context) {})
	router.PUT("/example", func(c *Context) {})
	router.DELETE("/example", func(c *Context) {})
	router.PATCH("/example", func(c *Context) {})
	router.HEAD("/example", func(c *Context) {})
	router.OPTIONS("/example", func(c *Context) {})

	PerformRequest(router, http.MethodGet, "/example?a=100")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	// I wrote these first (extending the above) but then realized they are more
	// like integration tests because they test the whole logging process rather
	// than individual functions. I'm not sure where these should go.
	buffer.Reset()
	PerformRequest(router, http.MethodPost, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodPost)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodPut, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodPut)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodDelete, "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodDelete)
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "PATCH", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PATCH")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(router, http.MethodGet, "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/notfound")
	logritelog.Default().Printf("<<< Exiting TestLoggerWithConfig(t=%v)", t)
}

func TestLoggerWithFormatter(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithFormatter(t=%v)", t)
	buffer := new(strings.Builder)

	d := DefaultWriter
	logritelog.Default().Printf("TestLoggerWithFormatter(t=%v): DefaultWriter → %v", t, DefaultWriter)
	logritelog.Default().Printf("<<< Exiting TestLoggerWithFormatter(t=%v)", t)
	DefaultWriter = buffer
	defer func() {
		DefaultWriter = d
	}()

	router := New()
	router.Use(LoggerWithFormatter(func(param LogFormatterParams) string {
		return fmt.Sprintf("[FORMATTER TEST] %v | %3d | %13v | %15s | %-7s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	}))
	router.GET("/example", func(c *Context) {})
	PerformRequest(router, http.MethodGet, "/example?a=100")

	// output test
	assert.Contains(t, buffer.String(), "[FORMATTER TEST]")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")
}

func TestLoggerWithConfigFormatting(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithConfigFormatting(t=%v)", t)
	var gotParam LogFormatterParams
	var gotKeys map[any]any
	buffer := new(strings.Builder)

	router := New()
	router.engine.trustedCIDRs, _ = router.engine.prepareTrustedCIDRs()

	router.Use(LoggerWithConfig(LoggerConfig{
		Output: buffer,
		Formatter: func(param LogFormatterParams) string {
			// for assert test
			gotParam = param
			logritelog.Default().Printf("TestLoggerWithConfigFormatting(t=%v): gotParam → %v", t, gotParam)

			return fmt.Sprintf("[FORMATTER TEST] %v | %3d | %13v | %15s | %-7s %s\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		},
	}))
	router.GET("/example", func(c *Context) {
		// set dummy ClientIP
		c.Request.Header.Set("X-Forwarded-For", "20.20.20.20")
		gotKeys = c.Keys
		logritelog.Default().Printf("TestLoggerWithConfigFormatting(t=%v): gotKeys → %v", t, gotKeys)
		time.Sleep(time.Millisecond)
	})
	PerformRequest(router, http.MethodGet, "/example?a=100")

	// output test
	assert.Contains(t, buffer.String(), "[FORMATTER TEST]")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), http.MethodGet)
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	// LogFormatterParams test
	assert.NotNil(t, gotParam.Request)
	assert.NotEmpty(t, gotParam.TimeStamp)
	assert.Equal(t, 200, gotParam.StatusCode)
	assert.NotEmpty(t, gotParam.Latency)
	assert.Equal(t, "20.20.20.20", gotParam.ClientIP)
	assert.Equal(t, http.MethodGet, gotParam.Method)
	assert.Equal(t, "/example?a=100", gotParam.Path)
	assert.Empty(t, gotParam.ErrorMessage)
	assert.Equal(t, gotKeys, gotParam.Keys)
	logritelog.Default().Printf("<<< Exiting TestLoggerWithConfigFormatting(t=%v)", t)
}

func TestDefaultLogFormatter(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestDefaultLogFormatter(t=%v)", t)
	timeStamp := time.Unix(1544173902, 0).UTC()

	termFalseParam := LogFormatterParams{
		TimeStamp:    timeStamp,
		StatusCode:   200,
		Latency:      time.Second * 5,
		ClientIP:     "20.20.20.20",
		Method:       http.MethodGet,
		Path:         "/",
		ErrorMessage: "",
		isTerm:       false,
	}

	termTrueParam := LogFormatterParams{
		TimeStamp:    timeStamp,
		StatusCode:   200,
		Latency:      time.Second * 5,
		ClientIP:     "20.20.20.20",
		Method:       http.MethodGet,
		Path:         "/",
		ErrorMessage: "",
		isTerm:       true,
	}
	termTrueLongDurationParam := LogFormatterParams{
		TimeStamp:    timeStamp,
		StatusCode:   200,
		Latency:      time.Millisecond * 9876543210,
		ClientIP:     "20.20.20.20",
		Method:       http.MethodGet,
		Path:         "/",
		ErrorMessage: "",
		isTerm:       true,
	}

	termFalseLongDurationParam := LogFormatterParams{
		TimeStamp:    timeStamp,
		StatusCode:   200,
		Latency:      time.Millisecond * 9876543210,
		ClientIP:     "20.20.20.20",
		Method:       http.MethodGet,
		Path:         "/",
		ErrorMessage: "",
		isTerm:       false,
	}

	assert.Equal(t, "[GIN] 2018/12/07 - 09:11:42 | 200 |       5s |     20.20.20.20 | GET      \"/\"\n", defaultLogFormatter(termFalseParam))
	assert.Equal(t, "[GIN] 2018/12/07 - 09:11:42 | 200 | 2743h29m0s |     20.20.20.20 | GET      \"/\"\n", defaultLogFormatter(termFalseLongDurationParam))

	assert.Equal(t, "[GIN] 2018/12/07 - 09:11:42 |\x1b[97;42m 200 \x1b[0m|\x1b[97;41m       5s \x1b[0m|     20.20.20.20 |\x1b[97;44m GET     \x1b[0m \"/\"\n", defaultLogFormatter(termTrueParam))
	assert.Equal(t, "[GIN] 2018/12/07 - 09:11:42 |\x1b[97;42m 200 \x1b[0m|\x1b[97;41m 2743h29m0s \x1b[0m|     20.20.20.20 |\x1b[97;44m GET     \x1b[0m \"/\"\n", defaultLogFormatter(termTrueLongDurationParam))
	logritelog.Default().Printf("<<< Exiting TestDefaultLogFormatter(t=%v)", t)
}

func TestColorForMethod(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestColorForMethod(t=%v)", t)
	colorForMethod := func(method string) string {
		p := LogFormatterParams{
			Method: method,
		}
		return p.MethodColor()
	}

	assert.Equal(t, blue, colorForMethod(http.MethodGet), "get should be blue")
	assert.Equal(t, cyan, colorForMethod(http.MethodPost), "post should be cyan")
	assert.Equal(t, yellow, colorForMethod(http.MethodPut), "put should be yellow")
	assert.Equal(t, red, colorForMethod(http.MethodDelete), "delete should be red")
	assert.Equal(t, green, colorForMethod("PATCH"), "patch should be green")
	assert.Equal(t, magenta, colorForMethod("HEAD"), "head should be magenta")
	assert.Equal(t, white, colorForMethod("OPTIONS"), "options should be white")
	assert.Equal(t, reset, colorForMethod("TRACE"), "trace is not defined and should be the reset color")
	logritelog.Default().Printf("<<< Exiting TestColorForMethod(t=%v)", t)
}

func TestColorForStatus(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestColorForStatus(t=%v)", t)
	colorForStatus := func(code int) string {
		p := LogFormatterParams{
			StatusCode: code,
		}
		return p.StatusCodeColor()
	}

	assert.Equal(t, white, colorForStatus(http.StatusContinue), "1xx should be white")
	assert.Equal(t, green, colorForStatus(http.StatusOK), "2xx should be green")
	assert.Equal(t, white, colorForStatus(http.StatusMovedPermanently), "3xx should be white")
	assert.Equal(t, yellow, colorForStatus(http.StatusNotFound), "4xx should be yellow")
	assert.Equal(t, red, colorForStatus(2), "other things should be red")
	logritelog.Default().Printf("<<< Exiting TestColorForStatus(t=%v)", t)
}

func TestColorForLatency(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestColorForLatency(t=%v)", t)
	colorForLatency := func(latency time.Duration) string {
		p := LogFormatterParams{
			Latency: latency,
		}
		return p.LatencyColor()
	}

	assert.Equal(t, white, colorForLatency(time.Duration(0)), "0 should be white")
	assert.Equal(t, white, colorForLatency(time.Millisecond*20), "20ms should be white")
	assert.Equal(t, green, colorForLatency(time.Millisecond*150), "150ms should be green")
	assert.Equal(t, cyan, colorForLatency(time.Millisecond*250), "250ms should be cyan")
	assert.Equal(t, blue, colorForLatency(time.Millisecond*400), "400ms should be blue")
	assert.Equal(t, yellow, colorForLatency(time.Millisecond*600), "600ms should be yellow")
	assert.Equal(t, magenta, colorForLatency(time.Millisecond*1500), "1.5s should be magenta")
	assert.Equal(t, red, colorForLatency(time.Second*3), "other things should be red")
	logritelog.Default().Printf("<<< Exiting TestColorForLatency(t=%v)", t)
}

func TestResetColor(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestResetColor(t=%v)", t)
	p := LogFormatterParams{}
	assert.Equal(t, string([]byte{27, 91, 48, 109}), p.ResetColor())
	logritelog.Default().Printf("<<< Exiting TestResetColor(t=%v)", t)
}

func TestIsOutputColor(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestIsOutputColor(t=%v)", t)
	// test with isTerm flag true.
	p := LogFormatterParams{
		isTerm: true,
	}

	consoleColorMode = autoColor
	logritelog.Default().Printf("TestIsOutputColor(t=%v): consoleColorMode → %v", t, consoleColorMode)
	assert.True(t, p.IsOutputColor())

	ForceConsoleColor()
	assert.True(t, p.IsOutputColor())

	DisableConsoleColor()
	assert.False(t, p.IsOutputColor())

	// test with isTerm flag false.
	p = LogFormatterParams{
		isTerm: false,
	}

	consoleColorMode = autoColor
	assert.False(t, p.IsOutputColor())

	ForceConsoleColor()
	assert.True(t, p.IsOutputColor())

	DisableConsoleColor()
	assert.False(t, p.IsOutputColor())

	// reset console color mode.
	logritelog.Default().Printf("<<< Exiting TestIsOutputColor(t=%v)", t)
	consoleColorMode = autoColor
}

func TestErrorLogger(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestErrorLogger(t=%v)", t)
	router := New()
	router.Use(ErrorLogger())
	router.GET("/error", func(c *Context) {
		c.Error(errors.New("this is an error")) //nolint: errcheck
	})
	router.GET("/abort", func(c *Context) {
		c.AbortWithError(http.StatusUnauthorized, errors.New("no authorized")) //nolint: errcheck
	})
	router.GET("/print", func(c *Context) {
		c.Error(errors.New("this is an error")) //nolint: errcheck
		c.String(http.StatusInternalServerError, "hola!")
	})

	w := PerformRequest(router, http.MethodGet, "/error")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, "{\"error\":\"this is an error\"}", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/abort")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, "{\"error\":\"no authorized\"}", w.Body.String())

	w = PerformRequest(router, http.MethodGet, "/print")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "hola!{\"error\":\"this is an error\"}", w.Body.String())
	logritelog.Default().Printf("<<< Exiting TestErrorLogger(t=%v)", t)
}

func TestLoggerWithWriterSkippingPaths(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithWriterSkippingPaths(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithWriter(buffer, "/skipped"))
	router.GET("/logged", func(c *Context) {})
	router.GET("/skipped", func(c *Context) {})

	PerformRequest(router, http.MethodGet, "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(router, http.MethodGet, "/skipped")
	assert.Contains(t, buffer.String(), "")
	logritelog.Default().Printf("<<< Exiting TestLoggerWithWriterSkippingPaths(t=%v)", t)
}

func TestLoggerWithConfigSkippingPaths(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithConfigSkippingPaths(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithConfig(LoggerConfig{
		Output:    buffer,
		SkipPaths: []string{"/skipped"},
	}))
	router.GET("/logged", func(c *Context) {})
	router.GET("/skipped", func(c *Context) {})

	PerformRequest(router, http.MethodGet, "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(router, http.MethodGet, "/skipped")
	assert.Contains(t, buffer.String(), "")
	logritelog.Default().Printf("<<< Exiting TestLoggerWithConfigSkippingPaths(t=%v)", t)
}

func TestLoggerWithConfigSkipper(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithConfigSkipper(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithConfig(LoggerConfig{
		Output: buffer,
		Skip: func(c *Context) bool {
			return c.Writer.Status() == http.StatusNoContent
		},
	}))
	router.GET("/logged", func(c *Context) { c.Status(http.StatusOK) })
	router.GET("/skipped", func(c *Context) { c.Status(http.StatusNoContent) })

	PerformRequest(router, http.MethodGet, "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(router, http.MethodGet, "/skipped")
	assert.Contains(t, buffer.String(), "")
	logritelog.Default().Printf("<<< Exiting TestLoggerWithConfigSkipper(t=%v)", t)
}

func TestDisableConsoleColor(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestDisableConsoleColor(t=%v)", t)
	New()
	assert.Equal(t, autoColor, consoleColorMode)
	DisableConsoleColor()
	assert.Equal(t, disableColor, consoleColorMode)

	// reset console color mode.
	consoleColorMode = autoColor
	logritelog.Default().Printf("TestDisableConsoleColor(t=%v): consoleColorMode → %v", t, consoleColorMode)
logritelog.Default().Printf("<<< Exiting TestDisableConsoleColor(t=%v)", t)
}

func TestForceConsoleColor(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestForceConsoleColor(t=%v)", t)
	New()
	assert.Equal(t, autoColor, consoleColorMode)
	ForceConsoleColor()
	assert.Equal(t, forceColor, consoleColorMode)

	// reset console color mode.
	consoleColorMode = autoColor
	logritelog.Default().Printf("TestForceConsoleColor(t=%v): consoleColorMode → %v", t, consoleColorMode)
logritelog.Default().Printf("<<< Exiting TestForceConsoleColor(t=%v)", t)
}

func TestLoggerWithConfigSkipQueryString(t *testing.T) {
	logritelog.Default().Printf(">>> Entering TestLoggerWithConfigSkipQueryString(t=%v)", t)
	buffer := new(strings.Builder)
	router := New()
	router.Use(LoggerWithConfig(LoggerConfig{
		Output:          buffer,
		SkipQueryString: true,
	}))
	router.GET("/logged", func(c *Context) { c.Status(http.StatusOK) })

	PerformRequest(router, "GET", "/logged?a=21")
	assert.Contains(t, buffer.String(), "200")
	assert.NotContains(t, buffer.String(), "a=21")
	logritelog.Default().Printf("<<< Exiting TestLoggerWithConfigSkipQueryString(t=%v)", t)
}
