// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ginS
import logritelog "log"

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var engine = sync.OnceValue(func() *gin.Engine {
	return gin.Default()
})

// LoadHTMLGlob is a wrapper for Engine.LoadHTMLGlob.
func LoadHTMLGlob(pattern string) {
	logritelog.Default().Printf(">>> Entering LoadHTMLGlob(pattern=%v)", pattern)
	engine().LoadHTMLGlob(pattern)
	logritelog.Default().Printf("<<< Exiting LoadHTMLGlob(pattern=%v)", pattern)
}

// LoadHTMLFiles is a wrapper for Engine.LoadHTMLFiles.
func LoadHTMLFiles(files ...string) {
	logritelog.Default().Printf(">>> Entering LoadHTMLFiles(files=%v)", files)
	engine().LoadHTMLFiles(files...)
	logritelog.Default().Printf("<<< Exiting LoadHTMLFiles(files=%v)", files)
}

// LoadHTMLFS is a wrapper for Engine.LoadHTMLFS.
func LoadHTMLFS(fs http.FileSystem, patterns ...string) {
	logritelog.Default().Printf(">>> Entering LoadHTMLFS(fs=%v,patterns=%v)", fs, patterns)
	engine().LoadHTMLFS(fs, patterns...)
	logritelog.Default().Printf("<<< Exiting LoadHTMLFS(fs=%v,patterns=%v)", fs, patterns)
}

// SetHTMLTemplate is a wrapper for Engine.SetHTMLTemplate.
func SetHTMLTemplate(templ *template.Template) {
	logritelog.Default().Printf(">>> Entering SetHTMLTemplate(templ=%v)", templ)
	engine().SetHTMLTemplate(templ)
	logritelog.Default().Printf("<<< Exiting SetHTMLTemplate(templ=%v)", templ)
}

// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
func NoRoute(handlers ...gin.HandlerFunc) {
	logritelog.Default().Printf(">>> Entering NoRoute(handlers=%v)", handlers)
	engine().NoRoute(handlers...)
	logritelog.Default().Printf("<<< Exiting NoRoute(handlers=%v)", handlers)
}

// NoMethod is a wrapper for Engine.NoMethod.
func NoMethod(handlers ...gin.HandlerFunc) {
	logritelog.Default().Printf(">>> Entering NoMethod(handlers=%v)", handlers)
	engine().NoMethod(handlers...)
	logritelog.Default().Printf("<<< Exiting NoMethod(handlers=%v)", handlers)
}

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
func Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	logritelog.Default().Printf(">>> Entering Group(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().Group(relativePath, handlers...)
}

// Handle is a wrapper for Engine.Handle.
func Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering Handle(httpMethod=%v,relativePath=%v,handlers=%v)", httpMethod, relativePath, handlers)
	return engine().Handle(httpMethod, relativePath, handlers...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering POST(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().POST(relativePath, handlers...)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering GET(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().GET(relativePath, handlers...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering DELETE(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().DELETE(relativePath, handlers...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering PATCH(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().PATCH(relativePath, handlers...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering PUT(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().PUT(relativePath, handlers...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering OPTIONS(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().OPTIONS(relativePath, handlers...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering HEAD(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().HEAD(relativePath, handlers...)
}

// Any is a wrapper for Engine.Any.
func Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering Any(relativePath=%v,handlers=%v)", relativePath, handlers)
	return engine().Any(relativePath, handlers...)
}

// StaticFile is a wrapper for Engine.StaticFile.
func StaticFile(relativePath, filepath string) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering StaticFile(relativePath=%v,filepath=%v)", relativePath, filepath)
	return engine().StaticFile(relativePath, filepath)
}

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use :
//
//	router.Static("/static", "/var/www")
func Static(relativePath, root string) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering Static(relativePath=%v,root=%v)", relativePath, root)
	return engine().Static(relativePath, root)
}

// StaticFS is a wrapper for Engine.StaticFS.
func StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering StaticFS(relativePath=%v,fs=%v)", relativePath, fs)
	return engine().StaticFS(relativePath, fs)
}

// Use attaches a global middleware to the router. i.e. the middlewares attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func Use(middlewares ...gin.HandlerFunc) gin.IRoutes {
	logritelog.Default().Printf(">>> Entering Use(middlewares=%v)", middlewares)
	return engine().Use(middlewares...)
}

// Routes returns a slice of registered routes.
func Routes() gin.RoutesInfo {
	logritelog.Default().Printf(">>> Entering Routes()")
	return engine().Routes()
}

// Run attaches to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func Run(addr ...string) (err error) {
	logritelog.Default().Printf(">>> Entering Run(addr=%v)", addr)
	return engine().Run(addr...)
}

// RunTLS attaches to a http.Server and starts listening and serving HTTPS requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func RunTLS(addr, certFile, keyFile string) (err error) {
	logritelog.Default().Printf(">>> Entering RunTLS(addr=%v,certFile=%v,keyFile=%v)", addr, certFile, keyFile)
	return engine().RunTLS(addr, certFile, keyFile)
}

// RunUnix attaches to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (i.e. a file)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func RunUnix(file string) (err error) {
	logritelog.Default().Printf(">>> Entering RunUnix(file=%v)", file)
	return engine().RunUnix(file)
}

// RunFd attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified file descriptor.
// Note: the method will block the calling goroutine indefinitely unless an error happens.
func RunFd(fd int) (err error) {
	logritelog.Default().Printf(">>> Entering RunFd(fd=%v)", fd)
	return engine().RunFd(fd)
}
