// Copyright 2019 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReaderRenderNoHeaders(t *testing.T) {
	logritelog.Default().Printf("[ERROR] >>> Entering TestReaderRenderNoHeaders(t=%v)", t)
	content := "test"
	r := Reader{
		ContentLength: int64(len(content)),
		Reader:        strings.NewReader(content),
	}
	err := r.Render(httptest.NewRecorder())
	require.NoError(t, err)
	logritelog.Default().Printf("[WARN] <<< Exiting TestReaderRenderNoHeaders(t=%v)", t)
}
