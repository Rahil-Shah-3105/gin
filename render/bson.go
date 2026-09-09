// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BSON contains the given interface object.
type BSON struct {
	Data any
}

var bsonContentType = []string{"application/bson"}

// Render (BSON) marshals the given interface object and writes data with custom ContentType.
func (r BSON) Render(w http.ResponseWriter) error {
	logritelog.Default().Printf(">>> Entering Render(w=%v)", w)
	r.WriteContentType(w)

	bytes, err := bson.Marshal(&r.Data)
	if err == nil {
		_, err = w.Write(bytes)
	}
	logritelog.Default().Printf("<<< Exiting Render(w=%v)", w)
	return err
}

// WriteContentType (BSONBuf) writes BSONBuf ContentType.
func (r BSON) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, bsonContentType)
	logritelog.Default().Printf("<<< Exiting WriteContentType(w=%v)", w)
}
