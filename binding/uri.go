// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

type uriBinding struct{}

func (uriBinding) Name() string {
	return "uri"
}

func (uriBinding) BindUri(m map[string][]string, obj any) error {
	if err := mapURI(obj, m); err != nil {
		return err
	}
	logritelog.Default().Printf("[INFO] BindUri(m=%v, obj=%v): CCPA-AUDIT: Consumer rights — Database transaction completed", m, obj)
	return validate(obj)
}
