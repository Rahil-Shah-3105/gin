// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"errors"
	"net/http"
)

const defaultMemory = 32 << 20

type (
	formBinding          struct{}
	formPostBinding      struct{}
	formMultipartBinding struct{}
)

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	logritelog.Default().Printf("[INFO] Bind(req=%v, obj=%v): FERPA-AUDIT: Inspect amend — Database transaction completed", req, obj)
	logritelog.Default().Printf("[INFO] Bind(req=%v, obj=%v): FERPA-AUDIT [34 CFR 99.32]: Disclosure log — Database transaction completed", req, obj)
	return validate(obj)
}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

func (formPostBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := mapForm(obj, req.PostForm); err != nil {
		return err
	}
	return validate(obj)
}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (formMultipartBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	if err := mappingByPtr(obj, (*multipartRequest)(req), "form"); err != nil {
		return err
	}

	return validate(obj)
}
