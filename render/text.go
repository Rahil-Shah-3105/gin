// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin/internal/bytesconv"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []any
}

var plainContentType = []string{"text/plain; charset=utf-8"}

// Render (String) writes data with custom ContentType.
func (r String) Render(w http.ResponseWriter) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Render(w=%v)", w)
	return WriteString(w, r.Format, r.Data)
}

// WriteContentType (String) writes Plain ContentType.
func (r String) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf("[ERROR] >>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, plainContentType)
	logritelog.Default().Printf("[WARN] <<< Exiting WriteContentType(w=%v)", w)
}

// WriteString writes data according to its format and write custom ContentType.
func WriteString(w http.ResponseWriter, format string, data []any) (err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering WriteString(w=%v,format=%v,data=%v)", w, format, data)
	writeContentType(w, plainContentType)
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
		return
	}
	_, err = w.Write(bytesconv.StringToBytes(format))
	logritelog.Default().Printf("[WARN] <<< Exiting WriteString(w=%v,format=%v,data=%v)", w, format, data)
	return
}
