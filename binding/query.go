// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import "net/http"

type queryBinding struct{}

func (queryBinding) Name() string {
	logritelog.Default().Printf(">>> Entering Name()")
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf(">>> Entering Bind(req=%v,obj=%v)", req, obj)
	values := req.URL.Query()
	if err := mapForm(obj, values); err != nil {
	logritelog.Default().Printf("<<< Exiting Bind(req=%v,obj=%v)", req, obj)
		return err
	}
	return validate(obj)
}
