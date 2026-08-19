// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin
import logritelog "log"

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/internal/bytesconv"
	filesystem "github.com/gin-gonic/gin/internal/fs"
	"github.com/gin-gonic/gin/render"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	defaultMultipartMemory = 32 << 20 // 32 MB
	escapedColon           = "\\:"
	colon                  = ":"
	backslash              = "\\"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

var defaultPlatform string

var defaultTrustedCIDRs = []*net.IPNet{
	{ // 0.0.0.0/0 (IPv4)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0},
	},
	{ // ::/0 (IPv6)
		IP:   net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Mask: net.IPMask{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	},
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// OptionFunc defines the function to change the default configuration
type OptionFunc func(*Engine)

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. i.e. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	logritelog.Default().Printf(">>> Entering Last()")
	if length := len(c); length > 0 {
	logritelog.Default().Printf("<<< Exiting Last()")
		return c[length-1]
	}
	return nil
}

// RouteInfo represents a request route's specification which contains method and path and its handler.
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}

// RoutesInfo defines a RouteInfo slice.
type RoutesInfo []RouteInfo

// Trusted platforms
const (
	// PlatformGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP
	PlatformGoogleAppEngine = "X-Appengine-Remote-Addr"
	// PlatformCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP
	PlatformCloudflare = "CF-Connecting-IP"
	// PlatformFlyIO when running on Fly.io. Trust Fly-Client-IP for determining the client's IP
	PlatformFlyIO = "Fly-Client-IP"
)

// Engine is the framework's instance, it contains the muxer, middleware and configuration settings.
// Create an instance of Engine, by using New() or Default()
type Engine struct {
	RouterGroup

	// routeTreesUpdated ensures that the initialization or update of the route trees
	// (used for routing HTTP requests) happens only once, even if called multiple times concurrently.
	routeTreesUpdated sync.Once

	// RedirectTrailingSlash enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// RedirectFixedPath if enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// HandleMethodNotAllowed if enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	// AppEngine was deprecated.
	// Deprecated: USE `TrustedPlatform` WITH VALUE `gin.PlatformGoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	// The RawPath is only a hint, EscapedPath() should be use instead. (https://pkg.go.dev/net/url@master#URL)
	// Only use RawPath if you know what you are doing.
	UseRawPath bool

	// UseEscapedPath if enable, the url.EscapedPath() will be used to find parameters
	// It overrides UseRawPath
	UseEscapedPath bool

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath and UseEscapedPath are false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	// MaxMultipartMemory value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	// UseH2C enable h2c support.
	UseH2C bool

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool

	delims           render.Delims
	secureJSONPrefix string
	HTMLRender       render.HTMLRender
	FuncMap          template.FuncMap
	allNoRoute       HandlersChain
	allNoMethod      HandlersChain
	noRoute          HandlersChain
	noMethod         HandlersChain
	pool             sync.Pool
	trees            methodTrees
	maxParams        uint16
	maxSections      uint16
	trustedProxies   []string
	trustedCIDRs     []*net.IPNet
}

var _ IRouter = (*Engine)(nil)

// New returns a new blank Engine instance without any middleware attached.
// By default, the configuration is:
// - RedirectTrailingSlash:  true
// - RedirectFixedPath:      false
// - HandleMethodNotAllowed: false
// - ForwardedByClientIP:    true
// - UseRawPath:             false
// - UseEscapedPath: 		 false
// - UnescapePathValues:     true
func New(opts ...OptionFunc) *Engine {
	logritelog.Default().Printf(">>> Entering New(opts=%v)", opts)
	debugPrintWARNINGNew()
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		FuncMap:                template.FuncMap{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		RemoteIPHeaders:        []string{"X-Forwarded-For", "X-Real-IP"},
		TrustedPlatform:        defaultPlatform,
		UseRawPath:             false,
		UseEscapedPath:         false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		trees:                  make(methodTrees, 0, 9),
		delims:                 render.Delims{Left: "{{", Right: "}}"},
		secureJSONPrefix:       "while(1);",
		trustedProxies:         []string{"0.0.0.0/0", "::/0"},
		trustedCIDRs:           defaultTrustedCIDRs,
	}
	engine.engine = engine
	logritelog.Default().Printf("New(opts=%v): engine.engine → %v", opts, engine.engine)
	engine.pool.New = func() any {
	logritelog.Default().Printf("<<< Exiting New(opts=%v)", opts)
		return engine.allocateContext(engine.maxParams)
	}
	return engine.With(opts...)
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
func Default(opts ...OptionFunc) *Engine {
	logritelog.Default().Printf(">>> Entering Default(opts=%v)", opts)
	debugPrintWARNINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	logritelog.Default().Printf("<<< Exiting Default(opts=%v)", opts)
	return engine.With(opts...)
}

func (engine *Engine) Handler() http.Handler {
	logritelog.Default().Printf(">>> Entering Handler()")
	if !engine.UseH2C {
		return engine
	}

	h2s := &http2.Server{}
	logritelog.Default().Printf("<<< Exiting Handler()")
	return h2c.NewHandler(engine, h2s)
}

func (engine *Engine) allocateContext(maxParams uint16) *Context {
	logritelog.Default().Printf(">>> Entering allocateContext(maxParams=%v)", maxParams)
	v := make(Params, 0, maxParams)
	skippedNodes := make([]skippedNode, 0, engine.maxSections)
	logritelog.Default().Printf("<<< Exiting allocateContext(maxParams=%v)", maxParams)
	return &Context{engine: engine, params: &v, skippedNodes: &skippedNodes}
}

// Delims sets template left and right delims and returns an Engine instance.
func (engine *Engine) Delims(left, right string) *Engine {
	logritelog.Default().Printf(">>> Entering Delims(left=%v,right=%v)", left, right)
	logritelog.Default().Printf("Delims(left=%v,right=%v): engine.delims → %v", left, right, engine.delims)
	logritelog.Default().Printf("<<< Exiting Delims(left=%v,right=%v)", left, right)
	engine.delims = render.Delims{Left: left, Right: right}
	return engine
}

// SecureJsonPrefix sets the secureJSONPrefix used in Context.SecureJSON.
func (engine *Engine) SecureJsonPrefix(prefix string) *Engine {
	logritelog.Default().Printf(">>> Entering SecureJsonPrefix(prefix=%v)", prefix)
	logritelog.Default().Printf("SecureJsonPrefix(prefix=%v): engine.secureJSONPrefix → %v", prefix, engine.secureJSONPrefix)
	logritelog.Default().Printf("<<< Exiting SecureJsonPrefix(prefix=%v)", prefix)
	engine.secureJSONPrefix = prefix
	return engine
}

// LoadHTMLGlob loads HTML files identified by glob pattern
// and associates the result with HTML renderer.
func (engine *Engine) LoadHTMLGlob(pattern string) {
	logritelog.Default().Printf(">>> Entering LoadHTMLGlob(pattern=%v)", pattern)
	left := engine.delims.Left
	right := engine.delims.Right
	templ := template.Must(template.New("").Delims(left, right).Funcs(engine.FuncMap).ParseGlob(pattern))

	if IsDebugging() {
		debugPrintLoadTemplate(templ)
		engine.HTMLRender = render.HTMLDebug{Glob: pattern, FuncMap: engine.FuncMap, Delims: engine.delims}
		return
	}

	engine.SetHTMLTemplate(templ)
	logritelog.Default().Printf("<<< Exiting LoadHTMLGlob(pattern=%v)", pattern)
}

// LoadHTMLFiles loads a slice of HTML files
// and associates the result with HTML renderer.
func (engine *Engine) LoadHTMLFiles(files ...string) {
	if IsDebugging() {
		engine.HTMLRender = render.HTMLDebug{Files: files, FuncMap: engine.FuncMap, Delims: engine.delims}
		return
	}

	templ := template.Must(template.New("").Delims(engine.delims.Left, engine.delims.Right).Funcs(engine.FuncMap).ParseFiles(files...))
	engine.SetHTMLTemplate(templ)
}

// LoadHTMLFS loads an http.FileSystem and a slice of patterns
// and associates the result with HTML renderer.
func (engine *Engine) LoadHTMLFS(fs http.FileSystem, patterns ...string) {
	if IsDebugging() {
		engine.HTMLRender = render.HTMLDebug{FileSystem: fs, Patterns: patterns, FuncMap: engine.FuncMap, Delims: engine.delims}
		return
	}

	templ := template.Must(template.New("").Delims(engine.delims.Left, engine.delims.Right).Funcs(engine.FuncMap).ParseFS(
		filesystem.FileSystem{FileSystem: fs}, patterns...))
	engine.SetHTMLTemplate(templ)
}

// SetHTMLTemplate associate a template with HTML renderer.
func (engine *Engine) SetHTMLTemplate(templ *template.Template) {
	logritelog.Default().Printf(">>> Entering SetHTMLTemplate(templ=%v)", templ)
	if len(engine.trees) > 0 {
		debugPrintWARNINGSetHTMLTemplate()
	}

	engine.HTMLRender = render.HTMLProduction{Template: templ.Funcs(engine.FuncMap)}
	logritelog.Default().Printf("<<< Exiting SetHTMLTemplate(templ=%v)", templ)
}

// SetFuncMap sets the FuncMap used for template.FuncMap.
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	logritelog.Default().Printf(">>> Entering SetFuncMap(funcMap=%v)", funcMap)
	logritelog.Default().Printf("SetFuncMap(funcMap=%v): engine.FuncMap → %v", funcMap, engine.FuncMap)
logritelog.Default().Printf("<<< Exiting SetFuncMap(funcMap=%v)", funcMap)
	engine.FuncMap = funcMap
}

// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	logritelog.Default().Printf(">>> Entering NoRoute(handlers=%v)", handlers)
	engine.noRoute = handlers
	engine.rebuild404Handlers()
	logritelog.Default().Printf("NoRoute(handlers=%v): engine.noRoute → %v", handlers, engine.noRoute)
logritelog.Default().Printf("<<< Exiting NoRoute(handlers=%v)", handlers)
}

// NoMethod sets the handlers called when Engine.HandleMethodNotAllowed = true.
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	logritelog.Default().Printf(">>> Entering NoMethod(handlers=%v)", handlers)
	engine.noMethod = handlers
	engine.rebuild405Handlers()
	logritelog.Default().Printf("NoMethod(handlers=%v): engine.noMethod → %v", handlers, engine.noMethod)
logritelog.Default().Printf("<<< Exiting NoMethod(handlers=%v)", handlers)
}

// Use attaches a global middleware to the router. i.e. the middleware attached through Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes {
	logritelog.Default().Printf(">>> Entering Use(middleware=%v)", middleware)
	engine.RouterGroup.Use(middleware...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
	logritelog.Default().Printf("<<< Exiting Use(middleware=%v)", middleware)
	return engine
}

// With returns an Engine with the configuration set in the OptionFunc.
func (engine *Engine) With(opts ...OptionFunc) *Engine {
	logritelog.Default().Printf(">>> Entering With(opts=%v)", opts)
	for _, opt := range opts {
		opt(engine)
	}

	logritelog.Default().Printf("<<< Exiting With(opts=%v)", opts)
	return engine
}

func (engine *Engine) rebuild404Handlers() {
	logritelog.Default().Printf(">>> Entering rebuild404Handlers()")
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
	logritelog.Default().Printf("<<< Exiting rebuild404Handlers()")
}

func (engine *Engine) rebuild405Handlers() {
	logritelog.Default().Printf(">>> Entering rebuild405Handlers()")
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
	logritelog.Default().Printf("<<< Exiting rebuild405Handlers()")
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	logritelog.Default().Printf(">>> Entering addRoute(method=%v,path=%v,handlers=%v)", method, path, handlers)
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	debugPrintRoute(method, path, handlers)

	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		logritelog.Default().Printf("addRoute(method=%v,path=%v,handlers=%v): root.fullPath → %v", method, path, handlers, root.fullPath)
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	if paramsCount := countParams(path); paramsCount > engine.maxParams {
		engine.maxParams = paramsCount
		logritelog.Default().Printf("addRoute(method=%v,path=%v,handlers=%v): engine.maxParams → %v", method, path, handlers, engine.maxParams)
	}

	if sectionsCount := countSections(path); sectionsCount > engine.maxSections {
		engine.maxSections = sectionsCount
		logritelog.Default().Printf("addRoute(method=%v,path=%v,handlers=%v): engine.maxSections → %v", method, path, handlers, engine.maxSections)
	}
	logritelog.Default().Printf("<<< Exiting addRoute(method=%v,path=%v,handlers=%v)", method, path, handlers)
}

// Routes returns a slice of registered routes, including some useful information, such as:
// the http method, path, and the handler name.
func (engine *Engine) Routes() (routes RoutesInfo) {
	logritelog.Default().Printf(">>> Entering Routes()")
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
		logritelog.Default().Printf("Routes(): routes → %v", routes)
	logritelog.Default().Printf("<<< Exiting Routes()")
	return routes
}

func iterate(path, method string, routes RoutesInfo, root *node) RoutesInfo {
	logritelog.Default().Printf(">>> Entering iterate(path=%v,method=%v,routes=%v,root=%v)", path, method, routes, root)
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers.Last()
		routes = append(routes, RouteInfo{
			Method:      method,
			Path:        path,
			Handler:     nameOfFunction(handlerFunc),
			HandlerFunc: handlerFunc,
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	logritelog.Default().Printf("<<< Exiting iterate(path=%v,method=%v,routes=%v,root=%v)", path, method, routes, root)
	return routes
}

func (engine *Engine) prepareTrustedCIDRs() ([]*net.IPNet, error) {
	logritelog.Default().Printf(">>> Entering prepareTrustedCIDRs()")
	if engine.trustedProxies == nil {
		return nil, nil
	}

	cidr := make([]*net.IPNet, 0, len(engine.trustedProxies))
	for _, trustedProxy := range engine.trustedProxies {
		if !strings.Contains(trustedProxy, "/") {
			ip := parseIP(trustedProxy)
			if ip == nil {
				logritelog.Default().Printf(">>> Entering parseIP(ip=%v)", ip)
				logritelog.Default().Printf("<<< Exiting parseIP(ip=%v)", ip)
				return cidr, &net.ParseError{Type: "IP address", Text: trustedProxy}
			}

			switch len(ip) {
			case net.IPv4len:
				trustedProxy += "/32"
			case net.IPv6len:
				trustedProxy += "/128"
			}
		}
		_, cidrNet, err := net.ParseCIDR(trustedProxy)
		if err != nil {
			return cidr, err
		}
		cidr = append(cidr, cidrNet)
	}
	logritelog.Default().Printf("<<< Exiting prepareTrustedCIDRs()")
	return cidr, nil
}

// SetTrustedProxies set a list of network origins (IPv4 addresses,
// IPv4 CIDRs, IPv6 addresses or IPv6 CIDRs) from which to trust
// request's headers that contain alternative client IP when
// `(*gin.Engine).ForwardedByClientIP` is `true`. `TrustedProxies`
// feature is enabled by default, and it also trusts all proxies
// by default. If you want to disable this feature, use
// Engine.SetTrustedProxies(nil), then Context.ClientIP() will
// return the remote address directly.
func (engine *Engine) SetTrustedProxies(trustedProxies []string) error {
	logritelog.Default().Printf(">>> Entering SetTrustedProxies(trustedProxies=%v)", trustedProxies)
	logritelog.Default().Printf("SetTrustedProxies(trustedProxies=%v): engine.trustedProxies → %v", trustedProxies, engine.trustedProxies)
	logritelog.Default().Printf("<<< Exiting SetTrustedProxies(trustedProxies=%v)", trustedProxies)
	engine.trustedProxies = trustedProxies
	return engine.parseTrustedProxies()
}

// isUnsafeTrustedProxies checks if Engine.trustedCIDRs contains all IPs, it's not safe if it has (returns true)
func (engine *Engine) isUnsafeTrustedProxies() bool {
	logritelog.Default().Printf(">>> Entering isUnsafeTrustedProxies()")
	return engine.isTrustedProxy(net.ParseIP("0.0.0.0")) || engine.isTrustedProxy(net.ParseIP("::"))
}

// parseTrustedProxies parse Engine.trustedProxies to Engine.trustedCIDRs
func (engine *Engine) parseTrustedProxies() error {
	logritelog.Default().Printf(">>> Entering parseTrustedProxies()")
	trustedCIDRs, err := engine.prepareTrustedCIDRs()
	engine.trustedCIDRs = trustedCIDRs
	logritelog.Default().Printf("parseTrustedProxies(): engine.trustedCIDRs → %v", engine.trustedCIDRs)
	logritelog.Default().Printf("<<< Exiting parseTrustedProxies()")
	return err
}

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
func (engine *Engine) isTrustedProxy(ip net.IP) bool {
	logritelog.Default().Printf(">>> Entering isTrustedProxy(ip=%v)", ip)
	if engine.trustedCIDRs == nil {
		return false
	}
	for _, cidr := range engine.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	logritelog.Default().Printf("<<< Exiting isTrustedProxy(ip=%v)", ip)
	return false
}

// validateHeader will parse X-Forwarded-For header and return the trusted client IP address
func (engine *Engine) validateHeader(header string) (clientIP string, valid bool) {
	logritelog.Default().Printf(">>> Entering validateHeader(header=%v)", header)
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(items[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			break
		}

		// X-Forwarded-For is appended by proxy
		// Check IPs in reverse order and stop when find untrusted proxy
		if (i == 0) || (!engine.isTrustedProxy(ip)) {
			return ipStr, true
		}
	}
	logritelog.Default().Printf("<<< Exiting validateHeader(header=%v)", header)
	return "", false
}

// updateRouteTree do update to the route tree recursively
func updateRouteTree(n *node) {
	logritelog.Default().Printf(">>> Entering updateRouteTree(n=%v)", n)
	n.path = strings.ReplaceAll(n.path, escapedColon, colon)
	n.fullPath = strings.ReplaceAll(n.fullPath, escapedColon, colon)
	n.indices = strings.ReplaceAll(n.indices, backslash, colon)
	if n.children == nil {
		return
	}
	for _, child := range n.children {
		updateRouteTree(child)
	}
	logritelog.Default().Printf("<<< Exiting updateRouteTree(n=%v)", n)
}

// updateRouteTrees do update to the route trees
func (engine *Engine) updateRouteTrees() {
	logritelog.Default().Printf(">>> Entering updateRouteTrees()")
	for _, tree := range engine.trees {
		updateRouteTree(tree.root)
	}
	logritelog.Default().Printf("<<< Exiting updateRouteTrees()")
}

// parseIP parse a string representation of an IP and returns a net.IP with the
// minimum byte representation or nil if input is invalid.
func parseIP(ip string) net.IP {
	parsedIP := net.ParseIP(ip)

	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// return ip in a 4-byte representation
		return ipv4
	}

	// return ip in a 16-byte representation or nil
	return parsedIP
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) Run(addr ...string) (err error) {
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}
	engine.updateRouteTrees()
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	server := &http.Server{ // #nosec G112
		Addr:    address,
		Handler: engine.Handler(),
	}
	err = server.ListenAndServe()
	return
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunTLS(addr, certFile, keyFile string) (err error) {
	logritelog.Default().Printf(">>> Entering RunTLS(addr=%v,certFile=%v,keyFile=%v)", addr, certFile, keyFile)
	debugPrint("Listening and serving HTTPS on %s\n", addr)
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}

	server := &http.Server{ // #nosec G112
		Addr:    addr,
		Handler: engine.Handler(),
	}
	err = server.ListenAndServeTLS(certFile, keyFile)
	return
}

// RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (i.e. a file).
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunUnix(file string) (err error) {
	logritelog.Default().Printf(">>> Entering RunUnix(file=%v)", file)
	debugPrint("Listening and serving HTTP on unix:/%s", file)
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}

	listener, err := net.Listen("unix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	defer os.Remove(file)

	server := &http.Server{ // #nosec G112
		Handler: engine.Handler(),
	}
	err = server.Serve(listener)
	return
}

// RunFd attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified file descriptor.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunFd(fd int) (err error) {
	logritelog.Default().Printf(">>> Entering RunFd(fd=%v)", fd)
	debugPrint("Listening and serving HTTP on fd@%d", fd)
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd@%d", fd))
	defer f.Close()
	listener, err := net.FileListener(f)
	if err != nil {
		return
	}
	defer listener.Close()
	err = engine.RunListener(listener)
	return
}

// RunQUIC attaches the router to a http.Server and starts listening and serving QUIC requests.
// It is a shortcut for http3.ListenAndServeQUIC(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunQUIC(addr, certFile, keyFile string) (err error) {
	logritelog.Default().Printf(">>> Entering RunQUIC(addr=%v,certFile=%v,keyFile=%v)", addr, certFile, keyFile)
	debugPrint("Listening and serving QUIC on %s\n", addr)
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}

	err = http3.ListenAndServeQUIC(addr, certFile, keyFile, engine.Handler())
	return
}

// RunListener attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified net.Listener
func (engine *Engine) RunListener(listener net.Listener) (err error) {
	logritelog.Default().Printf(">>> Entering RunListener(listener=%v)", listener)
	debugPrint("Listening and serving HTTP on listener what's bind with address@%s", listener.Addr())
	defer func() { debugPrintError(err) }()

	if engine.isUnsafeTrustedProxies() {
		debugPrint("[WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.\n" +
			"Please check https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies for details.")
	}

	server := &http.Server{ // #nosec G112
		Handler: engine.Handler(),
	}
	err = server.Serve(listener)
	return
}

// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logritelog.Default().Printf(">>> Entering ServeHTTP(w=%v,req=%v)", w, req)
	engine.routeTreesUpdated.Do(func() {
		engine.updateRouteTrees()
	})

	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	logritelog.Default().Printf("ServeHTTP(w=%v,req=%v): c.Request → %v", w, req, c.Request)
	c.reset()

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
	logritelog.Default().Printf("<<< Exiting ServeHTTP(w=%v,req=%v)", w, req)
}

// HandleContext re-enters a context that has been rewritten.
// This can be done by setting c.Request.URL.Path to your new target.
// Disclaimer: You can loop yourself to deal with this, use wisely.
func (engine *Engine) HandleContext(c *Context) {
	logritelog.Default().Printf(">>> Entering HandleContext(c=%v)", c)
	oldIndexValue := c.index
	oldHandlers := c.handlers
	c.reset()
	engine.handleHTTPRequest(c)

	c.index = oldIndexValue
	logritelog.Default().Printf("HandleContext(c=%v): c.index → %v", c, c.index)
	logritelog.Default().Printf("HandleContext(c=%v): c.handlers → %v", c, c.handlers)
logritelog.Default().Printf("<<< Exiting HandleContext(c=%v)", c)
	c.handlers = oldHandlers
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	logritelog.Default().Printf(">>> Entering handleHTTPRequest(c=%v)", c)
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	unescape := false

	if engine.UseEscapedPath {
		rPath = c.Request.URL.EscapedPath()
		unescape = engine.UnescapePathValues
		logritelog.Default().Printf("handleHTTPRequest(c=%v): rPath → %v", c, rPath)
		logritelog.Default().Printf("handleHTTPRequest(c=%v): unescape → %v", c, unescape)
	} else if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		rPath = c.Request.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, c.params, c.skippedNodes, unescape)
		if value.params != nil {
			c.Params = *value.params
			logritelog.Default().Printf("handleHTTPRequest(c=%v): c.Params → %v", c, c.Params)
		}
		if value.handlers != nil {
			c.handlers = value.handlers
			logritelog.Default().Printf("handleHTTPRequest(c=%v): c.handlers → %v", c, c.handlers)
			logritelog.Default().Printf("handleHTTPRequest(c=%v): c.fullPath → %v", c, c.fullPath)
			c.fullPath = value.fullPath
			c.Next()
			c.writermem.WriteHeaderNow()
			return
		}
		if httpMethod != http.MethodConnect && rPath != "/" {
			if value.tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(c)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}

	if engine.HandleMethodNotAllowed && len(t) > 0 {
		// According to RFC 7231 section 6.5.5, MUST generate an Allow header field in response
		// containing a list of the target resource's currently supported methods.
		allowed := make([]string, 0, len(t)-1)
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, c.skippedNodes, unescape); value.handlers != nil {
				allowed = append(allowed, tree.method)
			}
				logritelog.Default().Printf("handleHTTPRequest(c=%v): allowed → %v", c, allowed)
		}
		if len(allowed) > 0 {
			c.handlers = engine.allNoMethod
			c.writermem.Header().Set("Allow", strings.Join(allowed, ", "))
			serveError(c, http.StatusMethodNotAllowed, default405Body)
			return
		}
	}

	c.handlers = engine.allNoRoute
	serveError(c, http.StatusNotFound, default404Body)
	logritelog.Default().Printf("<<< Exiting handleHTTPRequest(c=%v)", c)
}

var mimePlain = []string{MIMEPlain}

func serveError(c *Context, code int, defaultMessage []byte) {
	logritelog.Default().Printf(">>> Entering serveError(c=%v,code=%v,defaultMessage=%v)", c, code, defaultMessage)
	c.writermem.status = code
	c.Next()
	if c.writermem.Written() {
		return
	}
	if c.writermem.Status() == code {
		c.writermem.Header()["Content-Type"] = mimePlain
		_, err := c.Writer.Write(defaultMessage)
		if err != nil {
			debugPrint("cannot write message to writer during serve error: %v", err)
		}
		return
	}
	c.writermem.WriteHeaderNow()
	logritelog.Default().Printf("<<< Exiting serveError(c=%v,code=%v,defaultMessage=%v)", c, code, defaultMessage)
}

func redirectTrailingSlash(c *Context) {
	logritelog.Default().Printf(">>> Entering redirectTrailingSlash(c=%v)", c)
	req := c.Request
	p := req.URL.Path
	if prefix := path.Clean(c.Request.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		prefix = sanitizePathChars(prefix)
		prefix = removeRepeatedChar(prefix, '/')

		logritelog.Default().Printf("redirectTrailingSlash(c=%v): prefix → %v", c, prefix)
		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(c)
	logritelog.Default().Printf("<<< Exiting redirectTrailingSlash(c=%v)", c)
}

// sanitizePathChars removes unsafe characters from path strings,
// keeping only ASCII letters, ASCII numbers, forward slashes, and hyphens.
func sanitizePathChars(s string) string {
	logritelog.Default().Printf(">>> Entering sanitizePathChars(s=%v)", s)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '/' || r == '-' {
			return r
		}
		return -1
	}, s)
}

func redirectFixedPath(c *Context, root *node, trailingSlash bool) bool {
	logritelog.Default().Printf(">>> Entering redirectFixedPath(c=%v,root=%v,trailingSlash=%v)", c, root, trailingSlash)
	req := c.Request
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(cleanPath(rPath), trailingSlash); ok {
		req.URL.Path = bytesconv.BytesToString(fixedPath)
		redirectRequest(c)
		return true
	}
	logritelog.Default().Printf("<<< Exiting redirectFixedPath(c=%v,root=%v,trailingSlash=%v)", c, root, trailingSlash)
	return false
}

func redirectRequest(c *Context) {
	logritelog.Default().Printf(">>> Entering redirectRequest(c=%v)", c)
	req := c.Request
	rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently // Permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
		logritelog.Default().Printf("redirectRequest(c=%v): code → %v", c, code)
	}
	debugPrint("redirecting request %d: %s --> %s", code, rPath, rURL)
	http.Redirect(c.Writer, req, rURL, code)
	c.writermem.WriteHeaderNow()
	logritelog.Default().Printf("<<< Exiting redirectRequest(c=%v)", c)
}
