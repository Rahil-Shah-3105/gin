// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"net/http"

	"google.golang.org/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Data any
}

var protobufContentType = []string{"application/x-protobuf"}

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (r ProtoBuf) Render(w http.ResponseWriter) error {
	logritelog.Default().Printf(">>> Entering Render(w=%v)", w)
	r.WriteContentType(w)

	bytes, err := proto.Marshal(r.Data.(proto.Message))
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in Render(w=%v): %v", w, err)
		return err
	}

	_, err = w.Write(bytes)
	logritelog.Default().Printf("<<< Exiting Render(w=%v)", w)
	return err
}

// WriteContentType (ProtoBuf) writes ProtoBuf ContentType.
func (r ProtoBuf) WriteContentType(w http.ResponseWriter) {
	logritelog.Default().Printf(">>> Entering WriteContentType(w=%v)", w)
	writeContentType(w, protobufContentType)
	logritelog.Default().Printf("<<< Exiting WriteContentType(w=%v)", w)
}
