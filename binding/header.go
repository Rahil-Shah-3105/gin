// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"net/http"
	"net/textproto"
	"reflect"
)

type headerBinding struct{}

func (headerBinding) Name() string {
	logritelog.Default().Printf(">>> Entering Name()")
	return "header"
}

func (headerBinding) Bind(req *http.Request, obj any) error {
	logritelog.Default().Printf(">>> Entering Bind(req=%v,obj=%v)", req, obj)
	if err := mapHeader(obj, req.Header); err != nil {
	logritelog.Default().Printf("<<< Exiting Bind(req=%v,obj=%v)", req, obj)
		return err
	}

	return validate(obj)
}

func mapHeader(ptr any, h map[string][]string) error {
	logritelog.Default().Printf(">>> Entering mapHeader(ptr=%v,h=%v)", ptr, h)
	return mappingByPtr(ptr, headerSource(h), "header")
}

type headerSource map[string][]string

var _ setter = headerSource(nil)

func (hs headerSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	logritelog.Default().Printf(">>> Entering TrySet(value=%v,field=%v,tagValue=%v,opt=%v)", value, field, tagValue, opt)
	return setByForm(value, field, hs, textproto.CanonicalMIMEHeaderKey(tagValue), opt)
}
