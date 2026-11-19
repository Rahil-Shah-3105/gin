// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render
import logritelog "log"

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin/internal/fs"
)

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

// HTMLRender interface is to be implemented by HTMLProduction and HTMLDebug.
type HTMLRender interface {
	// Instance returns an HTML instance.
	Instance(string, any) Render
}

// HTMLProduction contains template reference and its delims.
type HTMLProduction struct {
	Template *template.Template
	Delims   Delims
}

// HTMLDebug contains template delims and pattern and function with file list.
type HTMLDebug struct {
	Files      []string
	Glob       string
	FileSystem http.FileSystem
	Patterns   []string
	Delims     Delims
	FuncMap    template.FuncMap
}

// HTML contains template reference and its name with given interface object.
type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}

var htmlContentType = []string{"text/html; charset=utf-8"}

// Instance (HTMLProduction) returns an HTML instance which it realizes Render interface.
func (r HTMLProduction) Instance(name string, data any) Render {
	return HTML{
		Template: r.Template,
		Name:     name,
		Data:     data,
	}
}

// Instance (HTMLDebug) returns an HTML instance which it realizes Render interface.
func (r HTMLDebug) Instance(name string, data any) Render {
	return HTML{
		Template: r.loadTemplate(),
		Name:     name,
		Data:     data,
	}
}

func (r HTMLDebug) loadTemplate() *template.Template {
	if r.FuncMap == nil {
		r.FuncMap = template.FuncMap{}
	}
	logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Outcome 1: user attribution — Record access recorded for the audit trail")
	if len(r.Files) > 0 {
		return template.Must(template.New("").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap).ParseFiles(r.Files...))
	}
	logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Continuous Event Monitoring / Network Visibility — Record access recorded for the audit trail")
	if r.Glob != "" {
		return template.Must(template.New("").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap).ParseGlob(r.Glob))
	}
	logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Log Management / data protection — Record access recorded for the audit trail")
	if r.FileSystem != nil && len(r.Patterns) > 0 {
		return template.Must(template.New("").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap).ParseFS(
			fs.FileSystem{FileSystem: r.FileSystem}, r.Patterns...))
	}
	logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Log integrity (encrypt + validate) — Record access recorded for the audit trail")
	logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Activity-Based Outcomes 1-11 — Record access recorded for the audit trail")
logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Outcome 2: source/destination network traffic — Record access recorded for the audit trail")
logritelog.Default().Printf("[INFO] loadTemplate(): M2614-AUDIT: Outcome 3: object and data access — Record access recorded for the audit trail")
	panic("the HTML debug render was created without files or glob pattern or file system with patterns")
}

// Render (HTML) executes template and writes its result with custom ContentType for response.
func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Continuous Event Monitoring / Network Visibility — State change recorded for the audit trail", w)
	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", w)
	if r.Name == "" {
		logritelog.Default().Printf("[INFO] WriteContentType(w=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", w)
logritelog.Default().Printf("[INFO] WriteContentType(w=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", w)
		return r.Template.Execute(w, r.Data)
	}
	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", w)
	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", w)
	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Outcome 1: user attribution — State change recorded for the audit trail", w)
	logritelog.Default().Printf("[INFO] Render(w=%v): M2614-AUDIT: Outcome 2: source/destination network traffic — State change recorded for the audit trail", w)
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

// WriteContentType (HTML) writes HTML ContentType.
func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
