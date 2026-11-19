// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

type uriBinding struct{}

func (uriBinding) Name() string {
	logritelog.Default().Printf("[ERROR] >>> Entering Name()")
	return "uri"
}

func (uriBinding) BindUri(m map[string][]string, obj any) error {
	logritelog.Default().Printf("[ERROR] >>> Entering BindUri(m=%v,obj=%v)", m, obj)
	if err := mapURI(obj, m); err != nil {
		logritelog.Default().Printf("[ERROR] Exception in BindUri(m=%v,obj=%v): %v", m, obj, err)
		return err
	}
	logritelog.Default().Printf("[WARN] <<< Exiting BindUri(m=%v,obj=%v)", m, obj)
	return validate(obj)
}
