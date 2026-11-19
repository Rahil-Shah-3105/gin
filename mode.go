// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"flag"
	"io"
	"os"
	"sync/atomic"

	"github.com/gin-gonic/gin/binding"
)

// EnvGinMode indicates environment name for gin mode.
const EnvGinMode = "GIN_MODE"

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = "release"
	// TestMode indicates gin mode is test.
	TestMode = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

// DefaultWriter is the default io.Writer used by Gin for debug output and
// middleware output like Logger() or Recovery().
// Note that both Logger and Recovery provides custom ways to configure their
// output io.Writer.
// To support coloring in Windows use:
//
//	import "github.com/mattn/go-colorable"
//	gin.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Gin to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

var (
	ginMode  int32 = debugCode
	modeName atomic.Value
)

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	mode := os.Getenv(EnvGinMode)
	SetMode(mode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

// SetMode sets gin mode according to input string.
func SetMode(value string) {
	logritelog.Default().Printf("[ERROR] >>> Entering SetMode(value=%v)", value)
	if value == "" {
		if flag.Lookup("test.v") != nil {
			value = TestMode
		} else {
			value = DebugMode
		}
	}

	switch value {
	case DebugMode:
		atomic.StoreInt32(&ginMode, debugCode)
	case ReleaseMode:
		atomic.StoreInt32(&ginMode, releaseCode)
	case TestMode:
		atomic.StoreInt32(&ginMode, testCode)
	default:
		panic("gin mode unknown: " + value + " (available mode: debug release test)")
	}
	modeName.Store(value)
	logritelog.Default().Printf("[WARN] <<< Exiting SetMode(value=%v)", value)
}

// DisableBindValidation closes the default validator.
func DisableBindValidation() {
	logritelog.Default().Printf("[ERROR] >>> Entering DisableBindValidation()")
	binding.Validator = nil
	logritelog.Default().Printf("[DEBUG] DisableBindValidation(): binding.Validator → %v", binding.Validator)
logritelog.Default().Printf("[WARN] <<< Exiting DisableBindValidation()")
}

// EnableJsonDecoderUseNumber sets true for binding.EnableDecoderUseNumber to
// call the UseNumber method on the JSON Decoder instance.
func EnableJsonDecoderUseNumber() {
	logritelog.Default().Printf("[ERROR] >>> Entering EnableJsonDecoderUseNumber()")
	binding.EnableDecoderUseNumber = true
	logritelog.Default().Printf("[DEBUG] EnableJsonDecoderUseNumber(): binding.EnableDecoderUseNumber → %v", binding.EnableDecoderUseNumber)
logritelog.Default().Printf("[WARN] <<< Exiting EnableJsonDecoderUseNumber()")
}

// EnableJsonDecoderDisallowUnknownFields sets true for binding.EnableDecoderDisallowUnknownFields to
// call the DisallowUnknownFields method on the JSON Decoder instance.
func EnableJsonDecoderDisallowUnknownFields() {
	logritelog.Default().Printf("[ERROR] >>> Entering EnableJsonDecoderDisallowUnknownFields()")
	binding.EnableDecoderDisallowUnknownFields = true
	logritelog.Default().Printf("[DEBUG] EnableJsonDecoderDisallowUnknownFields(): binding.EnableDecoderDisallowUnknownFields → %v", binding.EnableDecoderDisallowUnknownFields)
logritelog.Default().Printf("[WARN] <<< Exiting EnableJsonDecoderDisallowUnknownFields()")
}

// Mode returns current gin mode.
func Mode() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Mode()")
	return modeName.Load().(string)
}
