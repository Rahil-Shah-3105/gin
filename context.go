// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
)

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEXML               = binding.MIMEXML
	MIMEXML2              = binding.MIMEXML2
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
	MIMEYAML              = binding.MIMEYAML
	MIMEYAML2             = binding.MIMEYAML2
	MIMETOML              = binding.MIMETOML
	MIMEPROTOBUF          = binding.MIMEPROTOBUF
	MIMEBSON              = binding.MIMEBSON
)

// BodyBytesKey indicates a default body bytes key.
const BodyBytesKey = "_gin-gonic/gin/bodybyteskey"

// ContextKey is the key that a Context returns itself for.
const ContextKey = "_gin-gonic/gin/contextkey"

type ContextKeyType int

const ContextRequestKey ContextKeyType = 0

// abortIndex represents a typical value used in abort functions.
const abortIndex int8 = math.MaxInt8 >> 1

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter

	Params   Params
	handlers HandlersChain
	index    int8
	fullPath string

	engine       *Engine
	params       *Params
	skippedNodes *[]skippedNode

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[any]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values

	// formCache caches c.Request.PostForm, which contains the parsed form data from POST, PATCH,
	// or PUT body parameters.
	formCache url.Values

	// SameSite allows a server to define a cookie attribute making it impossible for
	// the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (c *Context) reset() {
	logritelog.Default().Printf(">>> Entering reset()")
	c.Writer = &c.writermem
	c.Params = c.Params[:0]
	logritelog.Default().Printf("reset(): c.Params → %v", c.Params)
	logritelog.Default().Printf("reset(): c.handlers → %v", c.handlers)
	logritelog.Default().Printf("reset(): c.index → %v", c.index)
	logritelog.Default().Printf("reset(): c.Writer → %v", c.Writer)
	logritelog.Default().Printf("reset(): c.fullPath → %v", c.fullPath)
	logritelog.Default().Printf("reset(): c.Keys → %v", c.Keys)
	logritelog.Default().Printf("reset(): c.Errors → %v", c.Errors)
	logritelog.Default().Printf("reset(): c.Accepted → %v", c.Accepted)
	logritelog.Default().Printf("reset(): c.queryCache → %v", c.queryCache)
	logritelog.Default().Printf("reset(): c.formCache → %v", c.formCache)
	logritelog.Default().Printf("reset(): c.sameSite → %v", c.sameSite)
	c.handlers = nil
	c.index = -1

	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
	c.sameSite = 0
	*c.params = (*c.params)[:0]
	*c.skippedNodes = (*c.skippedNodes)[:0]
	logritelog.Default().Printf("<<< Exiting reset()")
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) Copy() *Context {
	logritelog.Default().Printf(">>> Entering Copy()")
	cp := Context{
		writermem: c.writermem,
		Request:   c.Request,
		engine:    c.engine,
	}

	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem
	logritelog.Default().Printf("Copy(): cp.Writer → %v", cp.Writer)
	logritelog.Default().Printf("Copy(): cp.index → %v", cp.index)
	logritelog.Default().Printf("Copy(): cp.handlers → %v", cp.handlers)
	logritelog.Default().Printf("Copy(): cp.fullPath → %v", cp.fullPath)
	cp.index = abortIndex
	cp.handlers = nil
	cp.fullPath = c.fullPath

	cKeys := c.Keys
	c.mu.RLock()
	cp.Keys = maps.Clone(cKeys)
	c.mu.RUnlock()

	cParams := c.Params
	cp.Params = make([]Param, len(cParams))
	copy(cp.Params, cParams)

	if c.Errors != nil {
		cp.Errors = make(errorMsgs, len(c.Errors))
		copy(cp.Errors, c.Errors)
	}

	if c.Accepted != nil {
		cp.Accepted = make([]string, len(c.Accepted))
		copy(cp.Accepted, c.Accepted)
	}

	logritelog.Default().Printf("<<< Exiting Copy()")
	return &cp
}

// HandlerName returns the main handler's name. For example if the handler is "handleGetUsers()",
// this function will return "main.handleGetUsers".
func (c *Context) HandlerName() string {
	logritelog.Default().Printf(">>> Entering HandlerName()")
	return nameOfFunction(c.handlers.Last())
}

// HandlerNames returns a list of all registered handlers for this context in descending order,
// following the semantics of HandlerName()
func (c *Context) HandlerNames() []string {
	logritelog.Default().Printf(">>> Entering HandlerNames()")
	hn := make([]string, 0, len(c.handlers))
	for _, val := range c.handlers {
		if val == nil {
			continue
		}
		hn = append(hn, nameOfFunction(val))
	}
	logritelog.Default().Printf("<<< Exiting HandlerNames()")
	return hn
}

// Handler returns the main handler.
func (c *Context) Handler() HandlerFunc {
	logritelog.Default().Printf(">>> Entering Handler()")
	return c.handlers.Last()
}

// FullPath returns a matched route full path. For not found routes
// returns an empty string.
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    c.FullPath() == "/user/:id" // true
//	})
func (c *Context) FullPath() string {
	logritelog.Default().Printf(">>> Entering FullPath()")
	return c.fullPath
}

/************************************/
/*********** FLOW CONTROL ***********/
/************************************/

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (c *Context) Next() {
	logritelog.Default().Printf(">>> Entering Next()")
	c.index++
	for c.index < safeInt8(len(c.handlers)) {
		if c.handlers[c.index] != nil {
			c.handlers[c.index](c)
		}
		c.index++
	}
	logritelog.Default().Printf("<<< Exiting Next()")
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	logritelog.Default().Printf(">>> Entering IsAborted()")
	return c.index >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *Context) Abort() {
	logritelog.Default().Printf(">>> Entering Abort()")
	logritelog.Default().Printf("Abort(): c.index → %v", c.index)
logritelog.Default().Printf("<<< Exiting Abort()")
	c.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	logritelog.Default().Printf(">>> Entering AbortWithStatus(code=%v)", code)
	c.Status(code)
	c.Writer.WriteHeaderNow()
	c.Abort()
	logritelog.Default().Printf("<<< Exiting AbortWithStatus(code=%v)", code)
}

// AbortWithStatusPureJSON calls `Abort()` and then `PureJSON` internally.
// This method stops the chain, writes the status code and return a JSON body without escaping.
// It also sets the Content-Type as "application/json".
func (c *Context) AbortWithStatusPureJSON(code int, jsonObj any) {
	logritelog.Default().Printf(">>> Entering AbortWithStatusPureJSON(code=%v,jsonObj=%v)", code, jsonObj)
	c.Abort()
	c.PureJSON(code, jsonObj)
	logritelog.Default().Printf("<<< Exiting AbortWithStatusPureJSON(code=%v,jsonObj=%v)", code, jsonObj)
}

// AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	logritelog.Default().Printf(">>> Entering AbortWithStatusJSON(code=%v,jsonObj=%v)", code, jsonObj)
	c.Abort()
	c.JSON(code, jsonObj)
	logritelog.Default().Printf("<<< Exiting AbortWithStatusJSON(code=%v,jsonObj=%v)", code, jsonObj)
}

// AbortWithError calls `AbortWithStatus()` and `Error()` internally.
// This method stops the chain, writes the status code and pushes the specified error to `c.Errors`.
// See Context.Error() for more details.
func (c *Context) AbortWithError(code int, err error) *Error {
	logritelog.Default().Printf(">>> Entering AbortWithError(code=%v,err=%v)", code, err)
	c.AbortWithStatus(code)
	logritelog.Default().Printf("<<< Exiting AbortWithError(code=%v,err=%v)", code, err)
	return c.Error(err)
	logritelog.Default().Printf(">>> Entering Error(err=%v)", err)
	logritelog.Default().Printf("Error(err=%v): parsedError → %v", err, parsedError)
	logritelog.Default().Printf("<<< Exiting Error(err=%v)", err)
}

/************************************/
/********* ERROR MANAGEMENT *********/
/************************************/

// Error attaches an error to the current context. The error is pushed to a list of errors.
// It's a good idea to call Error for each error that occurred during the resolution of a request.
// A middleware can be used to collect all the errors and push them to a database together,
// print a log, or append it in the HTTP response.
// Error will panic if err is nil.
func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}

	var parsedError *Error
	ok := errors.As(err, &parsedError)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes c.Keys if it was not used previously.
func (c *Context) Set(key any, value any) {
	logritelog.Default().Printf(">>> Entering Set(key=%v,value=%v)", key, value)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[any]any)
	}

	c.Keys[key] = value
	logritelog.Default().Printf("<<< Exiting Set(key=%v,value=%v)", key, value)
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key any) (value any, exists bool) {
	logritelog.Default().Printf(">>> Entering Get(key=%v)", key)
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key any) any {
	logritelog.Default().Printf(">>> Entering MustGet(key=%v)", key)
	if value, exists := c.Get(key); exists {
		return value
	}
	panic(fmt.Sprintf("key %v does not exist", key))
	logritelog.Default().Printf("<<< Exiting MustGet(key=%v)", key)
}

func getTyped[T any](c *Context, key any) (res T) {
	if val, ok := c.Get(key); ok && val != nil {
		res, _ = val.(T)
	}
	return
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key any) string {
	logritelog.Default().Printf(">>> Entering GetString(key=%v)", key)
	logritelog.Default().Printf(">>> Entering String(code=%v,format=%v,values=%v)", code, format, values)
	logritelog.Default().Printf("<<< Exiting String(code=%v,format=%v,values=%v)", code, format, values)
	return getTyped[string](c, key)
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key any) bool {
	logritelog.Default().Printf(">>> Entering GetBool(key=%v)", key)
	return getTyped[bool](c, key)
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key any) int {
	logritelog.Default().Printf(">>> Entering GetInt(key=%v)", key)
	return getTyped[int](c, key)
}

// GetInt8 returns the value associated with the key as an integer 8.
func (c *Context) GetInt8(key any) int8 {
	logritelog.Default().Printf(">>> Entering GetInt8(key=%v)", key)
	return getTyped[int8](c, key)
}

// GetInt16 returns the value associated with the key as an integer 16.
func (c *Context) GetInt16(key any) int16 {
	logritelog.Default().Printf(">>> Entering GetInt16(key=%v)", key)
	return getTyped[int16](c, key)
}

// GetInt32 returns the value associated with the key as an integer 32.
func (c *Context) GetInt32(key any) int32 {
	logritelog.Default().Printf(">>> Entering GetInt32(key=%v)", key)
	return getTyped[int32](c, key)
}

// GetInt64 returns the value associated with the key as an integer 64.
func (c *Context) GetInt64(key any) int64 {
	logritelog.Default().Printf(">>> Entering GetInt64(key=%v)", key)
	return getTyped[int64](c, key)
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key any) uint {
	logritelog.Default().Printf(">>> Entering GetUint(key=%v)", key)
	return getTyped[uint](c, key)
}

// GetUint8 returns the value associated with the key as an unsigned integer 8.
func (c *Context) GetUint8(key any) uint8 {
	logritelog.Default().Printf(">>> Entering GetUint8(key=%v)", key)
	return getTyped[uint8](c, key)
}

// GetUint16 returns the value associated with the key as an unsigned integer 16.
func (c *Context) GetUint16(key any) uint16 {
	logritelog.Default().Printf(">>> Entering GetUint16(key=%v)", key)
	return getTyped[uint16](c, key)
}

// GetUint32 returns the value associated with the key as an unsigned integer 32.
func (c *Context) GetUint32(key any) uint32 {
	logritelog.Default().Printf(">>> Entering GetUint32(key=%v)", key)
	return getTyped[uint32](c, key)
}

// GetUint64 returns the value associated with the key as an unsigned integer 64.
func (c *Context) GetUint64(key any) uint64 {
	logritelog.Default().Printf(">>> Entering GetUint64(key=%v)", key)
	return getTyped[uint64](c, key)
}

// GetFloat32 returns the value associated with the key as a float32.
func (c *Context) GetFloat32(key any) float32 {
	logritelog.Default().Printf(">>> Entering GetFloat32(key=%v)", key)
	return getTyped[float32](c, key)
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key any) float64 {
	logritelog.Default().Printf(">>> Entering GetFloat64(key=%v)", key)
	return getTyped[float64](c, key)
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key any) time.Time {
	logritelog.Default().Printf(">>> Entering GetTime(key=%v)", key)
	return getTyped[time.Time](c, key)
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key any) time.Duration {
	logritelog.Default().Printf(">>> Entering GetDuration(key=%v)", key)
	return getTyped[time.Duration](c, key)
}

// GetError returns the value associated with the key as an error.
func (c *Context) GetError(key any) error {
	logritelog.Default().Printf(">>> Entering GetError(key=%v)", key)
	return getTyped[error](c, key)
}

// GetIntSlice returns the value associated with the key as a slice of integers.
func (c *Context) GetIntSlice(key any) []int {
	logritelog.Default().Printf(">>> Entering GetIntSlice(key=%v)", key)
	return getTyped[[]int](c, key)
}

// GetInt8Slice returns the value associated with the key as a slice of int8 integers.
func (c *Context) GetInt8Slice(key any) []int8 {
	logritelog.Default().Printf(">>> Entering GetInt8Slice(key=%v)", key)
	return getTyped[[]int8](c, key)
}

// GetInt16Slice returns the value associated with the key as a slice of int16 integers.
func (c *Context) GetInt16Slice(key any) []int16 {
	logritelog.Default().Printf(">>> Entering GetInt16Slice(key=%v)", key)
	return getTyped[[]int16](c, key)
}

// GetInt32Slice returns the value associated with the key as a slice of int32 integers.
func (c *Context) GetInt32Slice(key any) []int32 {
	logritelog.Default().Printf(">>> Entering GetInt32Slice(key=%v)", key)
	return getTyped[[]int32](c, key)
}

// GetInt64Slice returns the value associated with the key as a slice of int64 integers.
func (c *Context) GetInt64Slice(key any) []int64 {
	logritelog.Default().Printf(">>> Entering GetInt64Slice(key=%v)", key)
	return getTyped[[]int64](c, key)
}

// GetUintSlice returns the value associated with the key as a slice of unsigned integers.
func (c *Context) GetUintSlice(key any) []uint {
	logritelog.Default().Printf(">>> Entering GetUintSlice(key=%v)", key)
	return getTyped[[]uint](c, key)
}

// GetUint8Slice returns the value associated with the key as a slice of uint8 integers.
func (c *Context) GetUint8Slice(key any) []uint8 {
	logritelog.Default().Printf(">>> Entering GetUint8Slice(key=%v)", key)
	return getTyped[[]uint8](c, key)
}

// GetUint16Slice returns the value associated with the key as a slice of uint16 integers.
func (c *Context) GetUint16Slice(key any) []uint16 {
	logritelog.Default().Printf(">>> Entering GetUint16Slice(key=%v)", key)
	return getTyped[[]uint16](c, key)
}

// GetUint32Slice returns the value associated with the key as a slice of uint32 integers.
func (c *Context) GetUint32Slice(key any) []uint32 {
	logritelog.Default().Printf(">>> Entering GetUint32Slice(key=%v)", key)
	return getTyped[[]uint32](c, key)
}

// GetUint64Slice returns the value associated with the key as a slice of uint64 integers.
func (c *Context) GetUint64Slice(key any) []uint64 {
	logritelog.Default().Printf(">>> Entering GetUint64Slice(key=%v)", key)
	return getTyped[[]uint64](c, key)
}

// GetFloat32Slice returns the value associated with the key as a slice of float32 numbers.
func (c *Context) GetFloat32Slice(key any) []float32 {
	logritelog.Default().Printf(">>> Entering GetFloat32Slice(key=%v)", key)
	return getTyped[[]float32](c, key)
}

// GetFloat64Slice returns the value associated with the key as a slice of float64 numbers.
func (c *Context) GetFloat64Slice(key any) []float64 {
	logritelog.Default().Printf(">>> Entering GetFloat64Slice(key=%v)", key)
	return getTyped[[]float64](c, key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key any) []string {
	logritelog.Default().Printf(">>> Entering GetStringSlice(key=%v)", key)
	return getTyped[[]string](c, key)
}

// GetErrorSlice returns the value associated with the key as a slice of errors.
func (c *Context) GetErrorSlice(key any) []error {
	logritelog.Default().Printf(">>> Entering GetErrorSlice(key=%v)", key)
	return getTyped[[]error](c, key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key any) map[string]any {
	logritelog.Default().Printf(">>> Entering GetStringMap(key=%v)", key)
	return getTyped[map[string]any](c, key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key any) map[string]string {
	logritelog.Default().Printf(">>> Entering GetStringMapString(key=%v)", key)
	return getTyped[map[string]string](c, key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key any) map[string][]string {
	logritelog.Default().Printf(">>> Entering GetStringMapStringSlice(key=%v)", key)
	return getTyped[map[string][]string](c, key)
}

// Delete deletes the key from the Context's Key map, if it exists.
// This operation is safe to be used by concurrent go-routines
func (c *Context) Delete(key any) {
	logritelog.Default().Printf(">>> Entering Delete(key=%v)", key)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys != nil {
		delete(c.Keys, key)
	}
	logritelog.Default().Printf("<<< Exiting Delete(key=%v)", key)
}

/************************************/
/************ INPUT DATA ************/
/************************************/

// Param returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    // a GET request to /user/john
//	    id := c.Param("id") // id == "john"
//	    // a GET request to /user/john/
//	    id := c.Param("id") // id == "/john/"
//	})
func (c *Context) Param(key string) string {
	logritelog.Default().Printf(">>> Entering Param(key=%v)", key)
	return c.Params.ByName(key)
}

// AddParam adds param to context and
// replaces path param key with given value for e2e testing purposes
// Example Route: "/user/:id"
// AddParam("id", 1)
// Result: "/user/1"
func (c *Context) AddParam(key, value string) {
	logritelog.Default().Printf(">>> Entering AddParam(key=%v,value=%v)", key, value)
	c.Params = append(c.Params, Param{Key: key, Value: value})
	logritelog.Default().Printf("<<< Exiting AddParam(key=%v,value=%v)", key, value)
}

// Query returns the keyed url query value if it exists,
// otherwise it returns an empty string `("")`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	    GET /path?id=1234&name=Manu&value=
//		   c.Query("id") == "1234"
//		   c.Query("name") == "Manu"
//		   c.Query("value") == ""
//		   c.Query("wtf") == ""
func (c *Context) Query(key string) (value string) {
	logritelog.Default().Printf(">>> Entering Query(key=%v)", key)
	value, _ = c.GetQuery(key)
	logritelog.Default().Printf("<<< Exiting Query(key=%v)", key)
	return
}

// DefaultQuery returns the keyed url query value if it exists,
// otherwise it returns the specified defaultValue string.
// See: Query() and GetQuery() for further information.
//
//	GET /?name=Manu&lastname=
//	c.DefaultQuery("name", "unknown") == "Manu"
//	c.DefaultQuery("id", "none") == "none"
//	c.DefaultQuery("lastname", "none") == ""
func (c *Context) DefaultQuery(key, defaultValue string) string {
	logritelog.Default().Printf(">>> Entering DefaultQuery(key=%v,defaultValue=%v)", key, defaultValue)
	if value, ok := c.GetQuery(key); ok {
	logritelog.Default().Printf("<<< Exiting DefaultQuery(key=%v,defaultValue=%v)", key, defaultValue)
		return value
	}
	return defaultValue
}

// GetQuery is like Query(), it returns the keyed url query value
// if it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns `("", false)`.
// It is shortcut for `c.Request.URL.Query().Get(key)`
//
//	GET /?name=Manu&lastname=
//	("Manu", true) == c.GetQuery("name")
//	("", false) == c.GetQuery("id")
//	("", true) == c.GetQuery("lastname")
func (c *Context) GetQuery(key string) (string, bool) {
	logritelog.Default().Printf(">>> Entering GetQuery(key=%v)", key)
	if values, ok := c.GetQueryArray(key); ok {
	logritelog.Default().Printf("<<< Exiting GetQuery(key=%v)", key)
		return values[0], ok
	}
	return "", false
}

// QueryArray returns a slice of strings for a given query key.
// The length of the slice depends on the number of params with the given key.
func (c *Context) QueryArray(key string) (values []string) {
	logritelog.Default().Printf(">>> Entering QueryArray(key=%v)", key)
	values, _ = c.GetQueryArray(key)
	logritelog.Default().Printf("<<< Exiting QueryArray(key=%v)", key)
	return
}

func (c *Context) initQueryCache() {
	logritelog.Default().Printf(">>> Entering initQueryCache()")
	if c.queryCache == nil {
		if c.Request != nil && c.Request.URL != nil {
			c.queryCache = c.Request.URL.Query()
		} else {
			c.queryCache = url.Values{}
			logritelog.Default().Printf("initQueryCache(): c.queryCache → %v", c.queryCache)
		}
	}
	logritelog.Default().Printf("<<< Exiting initQueryCache()")
}

// GetQueryArray returns a slice of strings for a given query key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	logritelog.Default().Printf(">>> Entering GetQueryArray(key=%v)", key)
	c.initQueryCache()
	values, ok = c.queryCache[key]
	logritelog.Default().Printf("<<< Exiting GetQueryArray(key=%v)", key)
	return
}

// QueryMap returns a map for a given query key.
func (c *Context) QueryMap(key string) (dicts map[string]string) {
	logritelog.Default().Printf(">>> Entering QueryMap(key=%v)", key)
	dicts, _ = c.GetQueryMap(key)
	logritelog.Default().Printf("<<< Exiting QueryMap(key=%v)", key)
	return
}

// GetQueryMap returns a map for a given query key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	logritelog.Default().Printf(">>> Entering GetQueryMap(key=%v)", key)
	c.initQueryCache()
	logritelog.Default().Printf("<<< Exiting GetQueryMap(key=%v)", key)
	return getMapFromFormData(c.queryCache, key)
}

// PostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns an empty string `("")`.
func (c *Context) PostForm(key string) (value string) {
	logritelog.Default().Printf(">>> Entering PostForm(key=%v)", key)
	value, _ = c.GetPostForm(key)
	logritelog.Default().Printf("<<< Exiting PostForm(key=%v)", key)
	return
}

// DefaultPostForm returns the specified key from a POST urlencoded form or multipart form
// when it exists, otherwise it returns the specified defaultValue string.
// See: PostForm() and GetPostForm() for further information.
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	logritelog.Default().Printf(">>> Entering DefaultPostForm(key=%v,defaultValue=%v)", key, defaultValue)
	if value, ok := c.GetPostForm(key); ok {
	logritelog.Default().Printf("<<< Exiting DefaultPostForm(key=%v,defaultValue=%v)", key, defaultValue)
		return value
	}
	return defaultValue
}

// GetPostForm is like PostForm(key). It returns the specified key from a POST urlencoded
// form or multipart form when it exists `(value, true)` (even when the value is an empty string),
// otherwise it returns ("", false).
// For example, during a PATCH request to update the user's email:
//
//	    email=mail@example.com  -->  ("mail@example.com", true) := GetPostForm("email") // set email to "mail@example.com"
//		email=                  -->  ("", true) := GetPostForm("email")                 // set email to ""
//	                            -->  ("", false) := GetPostForm("email")                // do nothing with email
func (c *Context) GetPostForm(key string) (string, bool) {
	logritelog.Default().Printf(">>> Entering GetPostForm(key=%v)", key)
	if values, ok := c.GetPostFormArray(key); ok {
	logritelog.Default().Printf("<<< Exiting GetPostForm(key=%v)", key)
		return values[0], ok
	}
	return "", false
}

// PostFormArray returns a slice of strings for a given form key.
// The length of the slice depends on the number of params with the given key.
func (c *Context) PostFormArray(key string) (values []string) {
	logritelog.Default().Printf(">>> Entering PostFormArray(key=%v)", key)
	values, _ = c.GetPostFormArray(key)
	logritelog.Default().Printf("<<< Exiting PostFormArray(key=%v)", key)
	return
}

func (c *Context) initFormCache() {
	logritelog.Default().Printf(">>> Entering initFormCache()")
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.Request
		if err := req.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				debugPrint("error on parse multipart form array: %v", err)
			}
		}
		c.formCache = req.PostForm
		logritelog.Default().Printf("initFormCache(): c.formCache → %v", c.formCache)
	}
	logritelog.Default().Printf("<<< Exiting initFormCache()")
}

// GetPostFormArray returns a slice of strings for a given form key, plus
// a boolean value whether at least one value exists for the given key.
func (c *Context) GetPostFormArray(key string) (values []string, ok bool) {
	logritelog.Default().Printf(">>> Entering GetPostFormArray(key=%v)", key)
	c.initFormCache()
	values, ok = c.formCache[key]
	logritelog.Default().Printf("<<< Exiting GetPostFormArray(key=%v)", key)
	return
}

// PostFormMap returns a map for a given form key.
func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	logritelog.Default().Printf(">>> Entering PostFormMap(key=%v)", key)
	dicts, _ = c.GetPostFormMap(key)
	logritelog.Default().Printf("<<< Exiting PostFormMap(key=%v)", key)
	return
}

// GetPostFormMap returns a map for a given form key, plus a boolean value
// whether at least one value exists for the given key.
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	logritelog.Default().Printf(">>> Entering GetPostFormMap(key=%v)", key)
	c.initFormCache()
	logritelog.Default().Printf("<<< Exiting GetPostFormMap(key=%v)", key)
	return getMapFromFormData(c.formCache, key)
}

// getMapFromFormData return a map which satisfies conditions.
// It parses from data with bracket notation like "key[subkey]=value" into a map.
func getMapFromFormData(m map[string][]string, key string) (map[string]string, bool) {
	logritelog.Default().Printf(">>> Entering getMapFromFormData(m=%v,key=%v)", m, key)
	d := make(map[string]string)
	found := false
	keyLen := len(key)

	for k, v := range m {
		if len(k) < keyLen+3 { // key + "[" + at least one char + "]"
			continue
		}

		if k[:keyLen] != key || k[keyLen] != '[' {
			continue
		}

		if j := strings.IndexByte(k[keyLen+1:], ']'); j > 0 {
			found = true
			logritelog.Default().Printf("getMapFromFormData(m=%v,key=%v): found → %v", m, key, found)
			d[k[keyLen+1:keyLen+1+j]] = v[0]
		}
	}

	logritelog.Default().Printf("<<< Exiting getMapFromFormData(m=%v,key=%v)", m, key)
	logritelog.Default().Printf(">>> Entering Data(code=%v,contentType=%v,data=%v)", code, contentType, data)
	logritelog.Default().Printf("<<< Exiting Data(code=%v,contentType=%v,data=%v)", code, contentType, data)
	return d, found
}

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	logritelog.Default().Printf(">>> Entering FormFile(name=%v)", name)
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	logritelog.Default().Printf("<<< Exiting FormFile(name=%v)", name)
	return fh, err
}

// MultipartForm is the parsed multipart form, including file uploads.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	logritelog.Default().Printf(">>> Entering MultipartForm()")
	err := c.Request.ParseMultipartForm(c.engine.MaxMultipartMemory)
	logritelog.Default().Printf("<<< Exiting MultipartForm()")
	return c.Request.MultipartForm, err
}

// SaveUploadedFile uploads the form file to specific dst.
// An optional perm argument specifies the permission bits used when creating
// the destination directory. If not provided, the default is 0750. The exact
// permission is enforced only on the destination directory and only when it is
// newly created by this call; pre-existing directories (e.g. /tmp) are not
// modified.
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string, perm ...fs.FileMode) error {
	logritelog.Default().Printf(">>> Entering SaveUploadedFile(file=%v,dst=%v,perm=%v)", file, dst, perm)
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	var mode os.FileMode = 0o750
	if len(perm) > 0 {
		mode = perm[0]
		logritelog.Default().Printf("SaveUploadedFile(file=%v,dst=%v,perm=%v): mode → %v", file, dst, perm, mode)
	}
	dir := filepath.Dir(dst)
	// Record whether the destination directory exists before MkdirAll, so we
	// only chmod a directory we just created. Chmod'ing a pre-existing directory
	// the process does not own (e.g. /tmp) fails with "operation not permitted"
	// (#4622). A non-ErrNotExist stat error also skips chmod and lets MkdirAll
	// surface the underlying failure.
	_, statErr := os.Stat(dir)
	if err = os.MkdirAll(dir, mode); err != nil {
		logritelog.Default().Printf("SaveUploadedFile(file=%v,dst=%v,perm=%v): err → %v", file, dst, perm, err)
		return err
	}
	if errors.Is(statErr, os.ErrNotExist) {
		if err = os.Chmod(dir, mode); err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	logritelog.Default().Printf("<<< Exiting SaveUploadedFile(file=%v,dst=%v,perm=%v)", file, dst, perm)
	return err
}

// Bind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body based on the Content-Type (e.g., JSON or XML).
// It decodes the payload into the struct specified as a pointer.
// It writes a 400 error and sets Content-Type header "text/plain" in the response if input is not valid.
func (c *Context) Bind(obj any) error {
	logritelog.Default().Printf(">>> Entering Bind(obj=%v)", obj)
	b := binding.Default(c.Request.Method, c.ContentType())
	logritelog.Default().Printf("<<< Exiting Bind(obj=%v)", obj)
	return c.MustBindWith(obj, b)
}

// BindJSON is a shortcut for c.MustBindWith(obj, binding.JSON).
func (c *Context) BindJSON(obj any) error {
	logritelog.Default().Printf(">>> Entering BindJSON(obj=%v)", obj)
	return c.MustBindWith(obj, binding.JSON)
}

// BindXML is a shortcut for c.MustBindWith(obj, binding.XML).
func (c *Context) BindXML(obj any) error {
	logritelog.Default().Printf(">>> Entering BindXML(obj=%v)", obj)
	logritelog.Default().Printf(">>> Entering XML(code=%v,obj=%v)", code, obj)
	logritelog.Default().Printf("<<< Exiting XML(code=%v,obj=%v)", code, obj)
	return c.MustBindWith(obj, binding.XML)
}

// BindQuery is a shortcut for c.MustBindWith(obj, binding.Query).
func (c *Context) BindQuery(obj any) error {
	logritelog.Default().Printf(">>> Entering BindQuery(obj=%v)", obj)
	return c.MustBindWith(obj, binding.Query)
}

// BindYAML is a shortcut for c.MustBindWith(obj, binding.YAML).
func (c *Context) BindYAML(obj any) error {
	logritelog.Default().Printf(">>> Entering BindYAML(obj=%v)", obj)
	logritelog.Default().Printf(">>> Entering YAML(code=%v,obj=%v)", code, obj)
	logritelog.Default().Printf("<<< Exiting YAML(code=%v,obj=%v)", code, obj)
	return c.MustBindWith(obj, binding.YAML)
}

// BindTOML is a shortcut for c.MustBindWith(obj, binding.TOML).
func (c *Context) BindTOML(obj any) error {
	logritelog.Default().Printf(">>> Entering BindTOML(obj=%v)", obj)
	logritelog.Default().Printf(">>> Entering TOML(code=%v,obj=%v)", code, obj)
	logritelog.Default().Printf("<<< Exiting TOML(code=%v,obj=%v)", code, obj)
	return c.MustBindWith(obj, binding.TOML)
}

// BindPlain is a shortcut for c.MustBindWith(obj, binding.Plain).
func (c *Context) BindPlain(obj any) error {
	logritelog.Default().Printf(">>> Entering BindPlain(obj=%v)", obj)
	return c.MustBindWith(obj, binding.Plain)
}

// BindHeader is a shortcut for c.MustBindWith(obj, binding.Header).
func (c *Context) BindHeader(obj any) error {
	logritelog.Default().Printf(">>> Entering BindHeader(obj=%v)", obj)
	logritelog.Default().Printf(">>> Entering Header(key=%v,value=%v)", key, value)
	logritelog.Default().Printf("<<< Exiting Header(key=%v,value=%v)", key, value)
	return c.MustBindWith(obj, binding.Header)
}

// BindUri binds the passed struct pointer using binding.Uri.
// It will abort the request with HTTP 400 if any error occurs.
func (c *Context) BindUri(obj any) error {
	logritelog.Default().Printf(">>> Entering BindUri(obj=%v)", obj)
	if err := c.ShouldBindUri(obj); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
	logritelog.Default().Printf("<<< Exiting BindUri(obj=%v)", obj)
		return err
	}
	return nil
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	logritelog.Default().Printf(">>> Entering MustBindWith(obj=%v,b=%v)", obj, b)
	err := c.ShouldBindWith(obj, b)
	if err != nil {
		var maxBytesErr *http.MaxBytesError

		// Note: When using sonic or go-json as JSON encoder, they do not propagate the http.MaxBytesError error
		// https://github.com/goccy/go-json/issues/485
		// https://github.com/bytedance/sonic/issues/800
		switch {
		case errors.As(err, &maxBytesErr):
			c.AbortWithError(http.StatusRequestEntityTooLarge, err).SetType(ErrorTypeBind) //nolint: errcheck
		default:
			c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
		}
		logritelog.Default().Printf(">>> Entering ShouldBindWith(obj=%v,b=%v)", obj, b)
	logritelog.Default().Printf("<<< Exiting MustBindWith(obj=%v,b=%v)", obj, b)
		return err
	}
	return nil
}

// ShouldBind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
//
// It parses the request's body based on the Content-Type (e.g., JSON or XML).
// It decodes the payload into the struct specified as a pointer.
// Like c.Bind() but this method does not set the response status code to 400 or abort if input is not valid.
func (c *Context) ShouldBind(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBind(obj=%v)", obj)
	b := binding.Default(c.Request.Method, c.ContentType())
	logritelog.Default().Printf("<<< Exiting ShouldBind(obj=%v)", obj)
	return c.ShouldBindWith(obj, b)
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
//
// Example:
//
//	POST /user
//	Content-Type: application/json
//
//	Request Body:
//	{
//		"name": "Manu",
//		"age": 20
//	}
//
//	type User struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//
//	var user User
//	if err := c.ShouldBindJSON(&user); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	c.JSON(http.StatusOK, user)
func (c *Context) ShouldBindJSON(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindJSON(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindXML is a shortcut for c.ShouldBindWith(obj, binding.XML).
// It works like ShouldBindJSON but binds the request body as XML data.
func (c *Context) ShouldBindXML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindXML(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.XML)
}

// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).
// It works like ShouldBindJSON but binds query parameters from the URL.
func (c *Context) ShouldBindQuery(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindQuery(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.Query)
}

// ShouldBindYAML is a shortcut for c.ShouldBindWith(obj, binding.YAML).
// It works like ShouldBindJSON but binds the request body as YAML data.
func (c *Context) ShouldBindYAML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindYAML(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.YAML)
}

// ShouldBindTOML is a shortcut for c.ShouldBindWith(obj, binding.TOML).
// It works like ShouldBindJSON but binds the request body as TOML data.
func (c *Context) ShouldBindTOML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindTOML(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.TOML)
}

// ShouldBindPlain is a shortcut for c.ShouldBindWith(obj, binding.Plain).
// It works like ShouldBindJSON but binds plain text data from the request body.
func (c *Context) ShouldBindPlain(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindPlain(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.Plain)
}

// ShouldBindHeader is a shortcut for c.ShouldBindWith(obj, binding.Header).
// It works like ShouldBindJSON but binds values from HTTP headers.
func (c *Context) ShouldBindHeader(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindHeader(obj=%v)", obj)
	return c.ShouldBindWith(obj, binding.Header)
}

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
// It works like ShouldBindJSON but binds parameters from the URI.
func (c *Context) ShouldBindUri(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindUri(obj=%v)", obj)
	m := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	logritelog.Default().Printf("<<< Exiting ShouldBindUri(obj=%v)", obj)
	return binding.Uri.BindUri(m, obj)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.Request, obj)
}

// ShouldBindBodyWith is similar with ShouldBindWith, but it stores the request
// body into the context, and reuse when it is called again.
//
// NOTE: This method reads the body before binding. So you should use
// ShouldBindWith for better performance if you need to call only once.
func (c *Context) ShouldBindBodyWith(obj any, bb binding.BindingBody) (err error) {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWith(obj=%v,bb=%v)", obj, bb)
	var body []byte
	if cb, ok := c.Get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in ShouldBindBodyWith(obj=%v,bb=%v): %v", obj, bb, err)
			return err
		}
		c.Set(BodyBytesKey, body)
	}
	logritelog.Default().Printf("<<< Exiting ShouldBindBodyWith(obj=%v,bb=%v)", obj, bb)
	return bb.BindBody(body, obj)
}

// ShouldBindBodyWithJSON is a shortcut for c.ShouldBindBodyWith(obj, binding.JSON).
func (c *Context) ShouldBindBodyWithJSON(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWithJSON(obj=%v)", obj)
	return c.ShouldBindBodyWith(obj, binding.JSON)
}

// ShouldBindBodyWithXML is a shortcut for c.ShouldBindBodyWith(obj, binding.XML).
func (c *Context) ShouldBindBodyWithXML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWithXML(obj=%v)", obj)
	return c.ShouldBindBodyWith(obj, binding.XML)
}

// ShouldBindBodyWithYAML is a shortcut for c.ShouldBindBodyWith(obj, binding.YAML).
func (c *Context) ShouldBindBodyWithYAML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWithYAML(obj=%v)", obj)
	return c.ShouldBindBodyWith(obj, binding.YAML)
}

// ShouldBindBodyWithTOML is a shortcut for c.ShouldBindBodyWith(obj, binding.TOML).
func (c *Context) ShouldBindBodyWithTOML(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWithTOML(obj=%v)", obj)
	return c.ShouldBindBodyWith(obj, binding.TOML)
}

// ShouldBindBodyWithPlain is a shortcut for c.ShouldBindBodyWith(obj, binding.Plain).
func (c *Context) ShouldBindBodyWithPlain(obj any) error {
	logritelog.Default().Printf(">>> Entering ShouldBindBodyWithPlain(obj=%v)", obj)
	return c.ShouldBindBodyWith(obj, binding.Plain)
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-IP]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (c *Context) ClientIP() string {
	logritelog.Default().Printf(">>> Entering ClientIP()")
	// Check if we're running on a trusted platform, continue running backwards if error
	if c.engine.TrustedPlatform != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := c.requestHeader(c.engine.TrustedPlatform); addr != "" {
			return addr
		}
	}

	// Legacy "AppEngine" flag
	if c.engine.AppEngine {
		log.Println(`The AppEngine flag is going to be deprecated. Please check issues #2723 and #2739 and use 'TrustedPlatform: gin.PlatformGoogleAppEngine' instead.`)
		if addr := c.requestHeader("X-Appengine-Remote-Addr"); addr != "" {
			return addr
		}
	}

	var (
		trusted  bool
		remoteIP net.IP
	)
	// If gin is listening a unix socket, always trust it.
	localAddr, ok := c.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if ok && strings.HasPrefix(localAddr.Network(), "unix") {
		trusted = true
		logritelog.Default().Printf("ClientIP(): trusted → %v", trusted)
	}

	// Fallback
	if !trusted {
		// It also checks if the remoteIP is a trusted proxy or not.
		// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
		// defined by Engine.SetTrustedProxies()
		remoteIP = net.ParseIP(c.RemoteIP())
		if remoteIP == nil {
			logritelog.Default().Printf(">>> Entering RemoteIP()")
			logritelog.Default().Printf("<<< Exiting RemoteIP()")
			return ""
		}
		trusted = c.engine.isTrustedProxy(remoteIP)
	}

	if trusted && c.engine.ForwardedByClientIP && c.engine.RemoteIPHeaders != nil {
		for _, headerName := range c.engine.RemoteIPHeaders {
			headerValue := strings.Join(c.Request.Header.Values(headerName), ",")
			ip, valid := c.engine.validateHeader(headerValue)
			if valid {
				return ip
			}
		}
	}
	logritelog.Default().Printf("<<< Exiting ClientIP()")
	return remoteIP.String()
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (c *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	logritelog.Default().Printf(">>> Entering ContentType()")
	return filterFlags(c.requestHeader("Content-Type"))
}

// IsWebsocket returns true if the request headers indicate that a websocket
// handshake is being initiated by the client.
func (c *Context) IsWebsocket() bool {
	logritelog.Default().Printf(">>> Entering IsWebsocket()")
	if strings.Contains(strings.ToLower(c.requestHeader("Connection")), "upgrade") &&
		strings.EqualFold(c.requestHeader("Upgrade"), "websocket") {
		logritelog.Default().Printf(">>> Entering requestHeader(key=%v)", key)
	logritelog.Default().Printf("<<< Exiting IsWebsocket()")
		return true
	}
	return false
}

// Scheme returns the HTTP scheme of the request ("http" or "https").
// When running behind reverse proxies or load balancers `Request.URL.Scheme` is usually empty.
// the original scheme is commonly forwarded via headers such as X-Forwarded-Proto.
// Reference:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/X-Forwarded-Proto
func (c *Context) Scheme() string {
	logritelog.Default().Printf(">>> Entering Scheme()")
	if c.Request.TLS != nil {
		return "https"
	}
	if scheme := c.requestHeader("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if scheme := c.requestHeader("X-Forwarded-Protocol"); scheme != "" {
		return scheme
	}
	if ssl := c.requestHeader("X-Forwarded-Ssl"); ssl == "on" {
		return "https"
	}
	if scheme := c.requestHeader("X-Url-Scheme"); scheme != "" {
		return scheme
	}
	if scheme := c.Request.URL.Scheme; scheme != "" {
	logritelog.Default().Printf("<<< Exiting Scheme()")
		return scheme
	}
	return "http"
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
// Uses http.StatusContinue constant for better code clarity.
func bodyAllowedForStatus(status int) bool {
	logritelog.Default().Printf(">>> Entering bodyAllowedForStatus(status=%v)", status)
	switch {
	case status >= http.StatusContinue && status < http.StatusOK:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
	logritelog.Default().Printf("<<< Exiting bodyAllowedForStatus(status=%v)", status)
		return false
	}
	return true
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	logritelog.Default().Printf(">>> Entering Status(code=%v)", code)
	c.Writer.WriteHeader(code)
	logritelog.Default().Printf("<<< Exiting Status(code=%v)", code)
}

// Header is an intelligent shortcut for c.Writer.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (c *Context) Header(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, value)
}

// GetHeader returns value from request headers.
func (c *Context) GetHeader(key string) string {
	logritelog.Default().Printf(">>> Entering GetHeader(key=%v)", key)
	return c.requestHeader(key)
}

// GetRawData returns stream data.
func (c *Context) GetRawData() ([]byte, error) {
	logritelog.Default().Printf(">>> Entering GetRawData()")
	if c.Request.Body == nil {
	logritelog.Default().Printf("<<< Exiting GetRawData()")
		return nil, errors.New("cannot read nil body")
	}
	return io.ReadAll(c.Request.Body)
}

// SetSameSite with cookie
func (c *Context) SetSameSite(samesite http.SameSite) {
	logritelog.Default().Printf(">>> Entering SetSameSite(samesite=%v)", samesite)
	logritelog.Default().Printf("SetSameSite(samesite=%v): c.sameSite → %v", samesite, c.sameSite)
logritelog.Default().Printf("<<< Exiting SetSameSite(samesite=%v)", samesite)
	c.sameSite = samesite
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	logritelog.Default().Printf(">>> Entering SetCookie(name=%v,value=%v,maxAge=%v,path=%v,domain=%v,secure=%v,httpOnly=%v)", name, value, maxAge, path, domain, secure, httpOnly)
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
	logritelog.Default().Printf("<<< Exiting SetCookie(name=%v,value=%v,maxAge=%v,path=%v,domain=%v,secure=%v,httpOnly=%v)", name, value, maxAge, path, domain, secure, httpOnly)
}

// SetCookieData adds a Set-Cookie header to the ResponseWriter's headers.
// It accepts a pointer to http.Cookie structure for more flexibility in setting cookie attributes.
// The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func (c *Context) SetCookieData(cookie *http.Cookie) {
	logritelog.Default().Printf(">>> Entering SetCookieData(cookie=<redacted>)")
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	logritelog.Default().Printf("SetCookieData(cookie=<redacted>): cookie.Path → %v", cookie.Path)
	if cookie.SameSite == http.SameSiteDefaultMode {
		cookie.SameSite = c.sameSite
		logritelog.Default().Printf("SetCookieData(cookie=<redacted>): cookie.SameSite → %v", cookie.SameSite)
	}
	http.SetCookie(c.Writer, cookie)
	logritelog.Default().Printf("<<< Exiting SetCookieData(cookie=<redacted>)")
}

// Cookie returns the named cookie provided in the request or
// ErrNoCookie if not found. And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (c *Context) Cookie(name string) (string, error) {
	logritelog.Default().Printf(">>> Entering Cookie(name=%v)", name)
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	logritelog.Default().Printf("<<< Exiting Cookie(name=%v)", name)
	return val, nil
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) {
	logritelog.Default().Printf(">>> Entering Render(code=%v,r=%v)", code, r)
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		// Pushing error to c.Errors
		_ = c.Error(err)
		c.Abort()
	}
	logritelog.Default().Printf("<<< Exiting Render(code=%v,r=%v)", code, r)
}

// HTML renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (c *Context) HTML(code int, name string, obj any) {
	logritelog.Default().Printf(">>> Entering HTML(code=%v,name=%v,obj=%v)", code, name, obj)
	instance := c.engine.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
	logritelog.Default().Printf("<<< Exiting HTML(code=%v,name=%v,obj=%v)", code, name, obj)
}

// IndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
// WARNING: we recommend using this only for development purposes since printing pretty JSON is
// more CPU and bandwidth consuming. Use Context.JSON() instead.
func (c *Context) IndentedJSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering IndentedJSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.IndentedJSON{Data: obj})
	logritelog.Default().Printf("<<< Exiting IndentedJSON(code=%v,obj=%v)", code, obj)
}

// SecureJSON serializes the given struct as Secure JSON into the response body.
// Default prepends "while(1)," to response body if the given struct is array values.
// It also sets the Content-Type as "application/json".
func (c *Context) SecureJSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering SecureJSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.SecureJSON{Prefix: c.engine.secureJSONPrefix, Data: obj})
	logritelog.Default().Printf("<<< Exiting SecureJSON(code=%v,obj=%v)", code, obj)
}

// JSONP serializes the given struct as JSON into the response body.
// It adds padding to response body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
func (c *Context) JSONP(code int, obj any) {
	logritelog.Default().Printf(">>> Entering JSONP(code=%v,obj=%v)", code, obj)
	callback := c.DefaultQuery("callback", "")
	if callback == "" {
		c.Render(code, render.JSON{Data: obj})
		return
	}
	c.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
	logritelog.Default().Printf("<<< Exiting JSONP(code=%v,obj=%v)", code, obj)
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering JSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.JSON{Data: obj})
	logritelog.Default().Printf("<<< Exiting JSON(code=%v,obj=%v)", code, obj)
}

// AsciiJSON serializes the given struct as JSON into the response body with unicode to ASCII string.
// It also sets the Content-Type as "application/json".
func (c *Context) AsciiJSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering AsciiJSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.AsciiJSON{Data: obj})
	logritelog.Default().Printf("<<< Exiting AsciiJSON(code=%v,obj=%v)", code, obj)
}

// PureJSON serializes the given struct as JSON into the response body.
// PureJSON, unlike JSON, does not replace special html characters with their unicode entities.
func (c *Context) PureJSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering PureJSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.PureJSON{Data: obj})
	logritelog.Default().Printf("<<< Exiting PureJSON(code=%v,obj=%v)", code, obj)
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) XML(code int, obj any) {
	c.Render(code, render.XML{Data: obj})
}

// PDF writes the given PDF binary data into the response body.
// It also sets the Content-Type as "application/pdf".
func (c *Context) PDF(code int, data []byte) {
	logritelog.Default().Printf(">>> Entering PDF(code=%v,data=%v)", code, data)
	c.Render(code, render.PDF{Data: data})
	logritelog.Default().Printf("<<< Exiting PDF(code=%v,data=%v)", code, data)
}

// YAML serializes the given struct as YAML into the response body.
func (c *Context) YAML(code int, obj any) {
	c.Render(code, render.YAML{Data: obj})
}

// TOML serializes the given struct as TOML into the response body.
func (c *Context) TOML(code int, obj any) {
	c.Render(code, render.TOML{Data: obj})
}

// ProtoBuf serializes the given struct as ProtoBuf into the response body.
func (c *Context) ProtoBuf(code int, obj any) {
	logritelog.Default().Printf(">>> Entering ProtoBuf(code=%v,obj=%v)", code, obj)
	c.Render(code, render.ProtoBuf{Data: obj})
	logritelog.Default().Printf("<<< Exiting ProtoBuf(code=%v,obj=%v)", code, obj)
}

// BSON serializes the given struct as BSON into the response body.
func (c *Context) BSON(code int, obj any) {
	logritelog.Default().Printf(">>> Entering BSON(code=%v,obj=%v)", code, obj)
	c.Render(code, render.BSON{Data: obj})
	logritelog.Default().Printf("<<< Exiting BSON(code=%v,obj=%v)", code, obj)
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// Redirect returns an HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
	logritelog.Default().Printf(">>> Entering Redirect(code=%v,location=%v)", code, location)
	c.Render(-1, render.Redirect{
		Code:     code,
		Location: location,
		Request:  c.Request,
	})
	logritelog.Default().Printf("<<< Exiting Redirect(code=%v,location=%v)", code, location)
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{
		ContentType: contentType,
		Data:        data,
	})
}

// DataFromReader writes the specified reader into the body stream and updates the HTTP code.
func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	logritelog.Default().Printf(">>> Entering DataFromReader(code=%v,contentLength=%v,contentType=%v,reader=%v,extraHeaders=%v)", code, contentLength, contentType, reader, extraHeaders)
	c.Render(code, render.Reader{
		Headers:       extraHeaders,
		ContentType:   contentType,
		ContentLength: contentLength,
		Reader:        reader,
	})
	logritelog.Default().Printf("<<< Exiting DataFromReader(code=%v,contentLength=%v,contentType=%v,reader=%v,extraHeaders=%v)", code, contentLength, contentType, reader, extraHeaders)
}

// File writes the specified file into the body stream in an efficient way.
func (c *Context) File(filepath string) {
	logritelog.Default().Printf(">>> Entering File(filepath=%v)", filepath)
	http.ServeFile(c.Writer, c.Request, filepath)
	logritelog.Default().Printf("<<< Exiting File(filepath=%v)", filepath)
}

// FileFromFS writes the specified file from http.FileSystem into the body stream in an efficient way.
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.Request.URL.Path = old
	}(c.Request.URL.Path)

	c.Request.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.Writer, c.Request)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	logritelog.Default().Printf(">>> Entering escapeQuotes(s=%v)", s)
	return quoteEscaper.Replace(s)
}

// FileAttachment writes the specified file into the body stream in an efficient way
// On the client side, the file will typically be downloaded with the given filename
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(filename)+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.Writer, c.Request, filepath)
}

// SSEvent writes a Server-Sent Event into the body stream.
func (c *Context) SSEvent(name string, message any) {
	logritelog.Default().Printf(">>> Entering SSEvent(name=%v,message=%v)", name, message)
	c.Render(-1, sse.Event{
		Event: name,
		Data:  message,
	})
	logritelog.Default().Printf("<<< Exiting SSEvent(name=%v,message=%v)", name, message)
}

// Stream sends a streaming response and returns a boolean
// indicates "Is client disconnected in middle of stream"
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	logritelog.Default().Printf(">>> Entering Stream(step=%v)", step)
	w := c.Writer
	clientGone := w.CloseNotify()
	for {
		select {
		case <-clientGone:
			return true
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				logritelog.Default().Printf("<<< Exiting Stream(step=%v)", step)
				return false
			}
		}
	}
}

/************************************/
/******** CONTENT NEGOTIATION *******/
/************************************/

// Negotiate contains all negotiations data.
type Negotiate struct {
	Offered      []string
	HTMLName     string
	HTMLData     any
	JSONData     any
	XMLData      any
	YAMLData     any
	Data         any
	TOMLData     any
	PROTOBUFData any
	BSONData     any
}

// Negotiate calls different Render according to acceptable Accept format.
func (c *Context) Negotiate(code int, config Negotiate) {
	logritelog.Default().Printf(">>> Entering Negotiate(code=%v,config=%v)", code, config)
	switch c.NegotiateFormat(config.Offered...) {
	case binding.MIMEJSON:
		data := chooseData(config.JSONData, config.Data)
		c.JSON(code, data)

	case binding.MIMEHTML:
		data := chooseData(config.HTMLData, config.Data)
		c.HTML(code, config.HTMLName, data)

	case binding.MIMEXML:
		data := chooseData(config.XMLData, config.Data)
		c.XML(code, data)

	case binding.MIMEYAML, binding.MIMEYAML2:
		data := chooseData(config.YAMLData, config.Data)
		c.YAML(code, data)

	case binding.MIMETOML:
		data := chooseData(config.TOMLData, config.Data)
		c.TOML(code, data)

	case binding.MIMEPROTOBUF:
		data := chooseData(config.PROTOBUFData, config.Data)
		c.ProtoBuf(code, data)

	case binding.MIMEBSON:
		data := chooseData(config.BSONData, config.Data)
		c.BSON(code, data)

	default:
		c.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server")) //nolint: errcheck
	}
	logritelog.Default().Printf("<<< Exiting Negotiate(code=%v,config=%v)", code, config)
}

// NegotiateFormat returns an acceptable Accept format.
func (c *Context) NegotiateFormat(offered ...string) string {
	logritelog.Default().Printf(">>> Entering NegotiateFormat(offered=%v)", offered)
	assert1(len(offered) > 0, "you must provide at least one offer")

	if c.Accepted == nil {
		c.Accepted = parseAccept(c.requestHeader("Accept"))
	}
	if len(c.Accepted) == 0 {
		return offered[0]
	}
	for _, accepted := range c.Accepted {
		for _, offer := range offered {
			// According to RFC 2616 and RFC 2396, non-ASCII characters are not allowed in headers,
			// therefore we can just iterate over the string without casting it into []rune
			i := 0
			for ; i < len(accepted) && i < len(offer); i++ {
				if accepted[i] == '*' || offer[i] == '*' {
					return offer
				}
				if accepted[i] != offer[i] {
					break
				}
			}
			if i == len(accepted) {
				return offer
			}
		}
	}
	logritelog.Default().Printf("<<< Exiting NegotiateFormat(offered=%v)", offered)
	return ""
}

// SetAccepted sets Accept header data.
func (c *Context) SetAccepted(formats ...string) {
	logritelog.Default().Printf(">>> Entering SetAccepted(formats=%v)", formats)
	logritelog.Default().Printf("SetAccepted(formats=%v): c.Accepted → %v", formats, c.Accepted)
logritelog.Default().Printf("<<< Exiting SetAccepted(formats=%v)", formats)
	c.Accepted = formats
}

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// hasRequestContext returns whether c.Request has Context and fallback.
func (c *Context) hasRequestContext() bool {
	logritelog.Default().Printf(">>> Entering hasRequestContext()")
	hasFallback := c.engine != nil && c.engine.ContextWithFallback
	hasRequestContext := c.Request != nil && c.Request.Context() != nil
	logritelog.Default().Printf("<<< Exiting hasRequestContext()")
	return hasFallback && hasRequestContext
}

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	logritelog.Default().Printf(">>> Entering Deadline()")
	if !c.hasRequestContext() {
	logritelog.Default().Printf("<<< Exiting Deadline()")
		return
	}
	return c.Request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *Context) Done() <-chan struct{} {
	logritelog.Default().Printf(">>> Entering Done()")
	if !c.hasRequestContext() {
	logritelog.Default().Printf("<<< Exiting Done()")
		return nil
	}
	return c.Request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (c *Context) Err() error {
	logritelog.Default().Printf(">>> Entering Err()")
	if !c.hasRequestContext() {
	logritelog.Default().Printf("<<< Exiting Err()")
		return nil
	}
	return c.Request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key any) any {
	logritelog.Default().Printf(">>> Entering Value(key=%v)", key)
	if key == ContextRequestKey {
		return c.Request
	}
	if key == ContextKey {
		return c
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if !c.hasRequestContext() {
	logritelog.Default().Printf("<<< Exiting Value(key=%v)", key)
		return nil
	}
	return c.Request.Context().Value(key)
}
