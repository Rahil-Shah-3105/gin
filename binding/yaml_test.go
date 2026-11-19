// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLBindingBindBody(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestYAMLBindingBindBody(t=%v)", t)
	var s struct {
		Foo string `yaml:"foo"`
	}
	err := yamlBinding{}.BindBody([]byte("foo: FOO"), &s)
	require.NoError(t, err)
	assert.Equal(t, "FOO", s.Foo)
	logritelog.Default().Printf("[WARN] <<< Exiting TestYAMLBindingBindBody(t=%v)", t)
}
