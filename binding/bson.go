// Copyright 2025 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type bsonBinding struct{}

func (bsonBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	logritelog.Default().Printf("[WARN] <<< Exiting Name()")
	return "bson"
}

func (b bsonBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering Bind(req=%v,obj=%v)", req, obj)
	buf, err := io.ReadAll(req.Body)
	if err == nil {
		err = b.BindBody(buf, obj)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting Bind(req=%v,obj=%v)", req, obj)
	return err
}

func (bsonBinding) BindBody(body []byte, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindBody(body=%v,obj=%v)", body, obj)
	logritelog.Default().Printf("[WARN] <<< Exiting BindBody(body=%v,obj=%v)", body, obj)
	return bson.Unmarshal(body, obj)
}
