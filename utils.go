// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"encoding/xml"
	"math"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

// BindKey indicates a default bind key.
const BindKey = "_gin-gonic/gin/bindkey"

// localhostIP indicates the default localhost IP address.
const localhostIP = "127.0.0.1"

// localhostIPv6 indicates the default localhost IPv6 address.
const localhostIPv6 = "::1"

// Bind is a helper function for given interface object and returns a Gin middleware.
func Bind(val any) HandlerFunc {
	logritelog.Default().Printf(">>> Entering Bind(val=%v)", val)
	value := reflect.ValueOf(val)
	if value.Kind() == reflect.Ptr {
		panic(`Bind struct can not be a pointer. Example:
	logritelog.Default().Printf("<<< Exiting Bind(val=%v)", val)
	Use: gin.Bind(Struct{}) instead of gin.Bind(&Struct{})
`)
	}
	typ := value.Type()

	return func(c *Context) {
		obj := reflect.New(typ).Interface()
		if c.Bind(obj) == nil {
			c.Set(BindKey, obj)
		}
	}
}

// WrapF is a helper function for wrapping http.HandlerFunc and returns a Gin middleware.
func WrapF(f http.HandlerFunc) HandlerFunc {
	logritelog.Default().Printf(">>> Entering WrapF(f=%v)", f)
	return func(c *Context) {
		f(c.Writer, c.Request)
	}
}

// WrapH is a helper function for wrapping http.Handler and returns a Gin middleware.
func WrapH(h http.Handler) HandlerFunc {
	logritelog.Default().Printf(">>> Entering WrapH(h=%v)", h)
	return func(c *Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// H is a shortcut for map[string]any
type H map[string]any

// MarshalXML allows type H to be used with xml.Marshal.
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	logritelog.Default().Printf(">>> Entering MarshalXML(e=%v,start=%v)", e, start)
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	logritelog.Default().Printf("MarshalXML(e=%v,start=%v): start.Name → %v", e, start, start.Name)
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	logritelog.Default().Printf("<<< Exiting MarshalXML(e=%v,start=%v)", e, start)
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func assert1(guard bool, text string) {
	logritelog.Default().Printf(">>> Entering assert1(guard=%v,text=%v)", guard, text)
	if !guard {
		panic(text)
	}
	logritelog.Default().Printf("<<< Exiting assert1(guard=%v,text=%v)", guard, text)
}

func filterFlags(content string) string {
	logritelog.Default().Printf(">>> Entering filterFlags(content=%v)", content)
	for i, char := range content {
		if char == ' ' || char == ';' {
		logritelog.Default().Printf("filterFlags(content=%v): char → %v", content, char)
			return content[:i]
		}
	}
	logritelog.Default().Printf("<<< Exiting filterFlags(content=%v)", content)
	return content
}

func chooseData(custom, wildcard any) any {
	logritelog.Default().Printf(">>> Entering chooseData(custom=%v,wildcard=%v)", custom, wildcard)
	if custom != nil {
		return custom
	}
	if wildcard != nil {
		return wildcard
	}
	panic("negotiation config is invalid")
	logritelog.Default().Printf("<<< Exiting chooseData(custom=%v,wildcard=%v)", custom, wildcard)
}

func parseAccept(acceptHeader string) []string {
	logritelog.Default().Printf(">>> Entering parseAccept(acceptHeader=%v)", acceptHeader)
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			logritelog.Default().Printf("parseAccept(acceptHeader=%v): part → %v", acceptHeader, part)
			logritelog.Default().Printf("parseAccept(acceptHeader=%v): out → %v", acceptHeader, out)
			out = append(out, part)
		}
	}
	logritelog.Default().Printf("<<< Exiting parseAccept(acceptHeader=%v)", acceptHeader)
	return out
}

func lastChar(str string) uint8 {
	logritelog.Default().Printf(">>> Entering lastChar(str=%v)", str)
	if str == "" {
		panic("The length of the string can't be 0")
	}
	logritelog.Default().Printf("<<< Exiting lastChar(str=%v)", str)
	return str[len(str)-1]
}

func nameOfFunction(f any) string {
	logritelog.Default().Printf(">>> Entering nameOfFunction(f=%v)", f)
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	logritelog.Default().Printf(">>> Entering joinPaths(absolutePath=%v,relativePath=%v)", absolutePath, relativePath)
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
	logritelog.Default().Printf("<<< Exiting joinPaths(absolutePath=%v,relativePath=%v)", absolutePath, relativePath)
		return finalPath + "/"
	}
	return finalPath
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	logritelog.Default().Printf(">>> Entering isASCII(s=%v)", s)
	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	logritelog.Default().Printf("<<< Exiting isASCII(s=%v)", s)
	return true
}

// safeInt8 converts int to int8 safely, capping at math.MaxInt8
func safeInt8(n int) int8 {
	logritelog.Default().Printf(">>> Entering safeInt8(n=%v)", n)
	if n > math.MaxInt8 {
	logritelog.Default().Printf("<<< Exiting safeInt8(n=%v)", n)
		return math.MaxInt8
	}
	return int8(n)
}

// safeUint16 converts int to uint16 safely, capping at math.MaxUint16
func safeUint16(n int) uint16 {
	logritelog.Default().Printf(">>> Entering safeUint16(n=%v)", n)
	if n > math.MaxUint16 {
	logritelog.Default().Printf("<<< Exiting safeUint16(n=%v)", n)
		return math.MaxUint16
	}
	return uint16(n)
}
