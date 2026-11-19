// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import "net/http"

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	values := req.URL.Query()
	if err := mapForm(obj, values); err != nil {
		return err
	}
	logritelog.Default().Printf("[INFO] Bind(req=%v, obj=%v): FERPA-AUDIT: Inspect amend — Database transaction completed", req, obj)
	logritelog.Default().Printf("[INFO] Bind(req=%v, obj=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — Database transaction completed", req, obj)
	return validate(obj)
}
