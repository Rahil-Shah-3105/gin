// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"os"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

func init() {
	logritelog.Default().Printf("[ERROR] >>> Entering init()")
	os.Setenv(EnvGinMode, TestMode)
	logritelog.Default().Printf("[WARN] <<< Exiting init()")
}

func TestSetMode(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestSetMode(t=%v)", t)
	assert.Equal(t, int32(testCode), atomic.LoadInt32(&ginMode))
	assert.Equal(t, TestMode, Mode())
	os.Unsetenv(EnvGinMode)

	SetMode("")
	assert.Equal(t, int32(testCode), atomic.LoadInt32(&ginMode))
	assert.Equal(t, TestMode, Mode())

	SetMode(DebugMode)
	assert.Equal(t, int32(debugCode), atomic.LoadInt32(&ginMode))
	assert.Equal(t, DebugMode, Mode())

	SetMode(ReleaseMode)
	assert.Equal(t, int32(releaseCode), atomic.LoadInt32(&ginMode))
	assert.Equal(t, ReleaseMode, Mode())

	SetMode(TestMode)
	assert.Equal(t, int32(testCode), atomic.LoadInt32(&ginMode))
	assert.Equal(t, TestMode, Mode())

	assert.Panics(t, func() { SetMode("unknown") })
	logritelog.Default().Printf("[WARN] <<< Exiting TestSetMode(t=%v)", t)
}

func TestDisableBindValidation(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestDisableBindValidation(t=%v)", t)
	v := binding.Validator
	assert.NotNil(t, binding.Validator)
	DisableBindValidation()
	assert.Nil(t, binding.Validator)
	binding.Validator = v
	logritelog.Default().Printf("[DEBUG] TestDisableBindValidation(t=%v): binding.Validator → %v", t, binding.Validator)
logritelog.Default().Printf("[WARN] <<< Exiting TestDisableBindValidation(t=%v)", t)
}

func TestEnableJsonDecoderUseNumber(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEnableJsonDecoderUseNumber(t=%v)", t)
	assert.False(t, binding.EnableDecoderUseNumber)
	EnableJsonDecoderUseNumber()
	assert.True(t, binding.EnableDecoderUseNumber)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEnableJsonDecoderUseNumber(t=%v)", t)
}

func TestEnableJsonDecoderDisallowUnknownFields(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestEnableJsonDecoderDisallowUnknownFields(t=%v)", t)
	assert.False(t, binding.EnableDecoderDisallowUnknownFields)
	EnableJsonDecoderDisallowUnknownFields()
	assert.True(t, binding.EnableDecoderDisallowUnknownFields)
	logritelog.Default().Printf("[WARN] <<< Exiting TestEnableJsonDecoderDisallowUnknownFields(t=%v)", t)
}
