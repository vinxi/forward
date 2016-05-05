// Package router implements a simple URL pattern muxer router
// with hierarchical middleware layer.
package router

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/vinxi/forward.v0"
	"gopkg.in/vinxi/layer.v0"
)

var (
	// ErrNoRouteMatch is returned as typed error when no route can be matched.
	ErrNoRouteMatch = errors.New("router: cannot match any route")
)

// Router is an HTTP request multiplexer. It matches the URL of each
// incoming request against a list of registered patterns with their associated
// methods and calls the handler for the pattern that most closely matches the
// URL.
//
// Pattern matching attempts each pattern in the order in which they were
// registered.
//
// Patterns may contain literals or captures. Capture names start with a colon
// and consist of letters A-Z, a-z, _, and 0-9. The rest of the pattern
// matches literally. The portion of the URL matching each name ends with an
// occurrence of the character in the pattern immediately following the name,
// or a /, whichever comes first. It is possible for a name to match the empty
// string.
//
// Example pattern with one capture:
//   /hello/:name
// Will match:
//   /hello/blake
//   /hello/keith
// Will not match:
//   /hello/blake/
//   /hello/blake/foo
//   /foo
//   /foo/bar
//
// Example 2:
//    /hello/:name/
// Will match:
//   /hello/blake/
//   /hello/keith/foo
//   /hello/blake
//   /hello/keith
// Will not match:
//   /foo
//   /foo/bar
//
// A pattern ending with a slash will add an implicit redirect for its non-slash
// version. For example: Get("/foo/", handler) also registers
// Get("/foo", handler) as a redirect. You may override it by registering
// Get("/foo", anotherhandler) before the slash version.
//
// Retrieve the capture from the r.URL.Query().Get(":name") in a handler (note
// the colon). If a capture name appears more than once, the additional values
// are appended to the previous values (see
// http://golang.org/pkg/net/url/#Values)
//
// A trivial example server is:
//
//	package main
//
//	import (
//		"io"
//		"net/http"
//		"github.com/bmizerany/pat"
//		"log"
//	)
//
//	// hello world, the web server
//	func HelloServer(w http.ResponseWriter, req *http.Request) {
//		io.WriteString(w, "hello, "+req.URL.Query().Get(":name")+"!\n")
//	}
//
//	func main() {
//		m := pat.New()
//		m.Get("/hello/:name", http.HandlerFunc(HelloServer))
//
//		// Register this pat with the default serve mux so that other packages
//		// may also be exported. (i.e. /debug/pprof/*)
//		http.Handle("/", m)
//		err := http.ListenAndServe(":12345", nil)
//		if err != nil {
//			log.Fatal("ListenAndServe: ", err)
//		}
//	}
//
// When "Method Not Allowed":
//
// Router knows what methods are allowed given a pattern and a URI. For
// convenience, Router will add the Allow header for requests that
// match a pattern for a method other than the method requested and set the
// Status to "405 Method Not Allowed".
//
// If the NotFound handler is set, then it is used whenever the pattern doesn't
// match the request path for the current method (and the Allow header is not
// altered).
type Router struct {
	// ForceMethodNotAllowed, if true, is used to return a 405 Method Not Allowed response
	// in case that the router has no configured routes for the incoming request method.
	// Defaults to false.
	ForceMethodNotAllowed bool

	// NotFound, if set, is used whenever the request doesn't match any
	// pattern for its method. NotFound should be set before serving any
	// requests.
	NotFound http.Handler

	// Layer provides middleware layer capabilities to the router.
	Layer *layer.Layer

	// Routes associates HTTP methods with path patterns handlers.
	Routes map[string][]*Route
}

// New creates a new Router.
func New() *Router {
	return &Router{Layer: layer.New(), Routes: make(map[string][]*Route)}
}

// Forward defines the default URL to forward incoming traffic.
func (r *Router) Forward(uri string) *Router {
	r.Layer.UseFinalHandler(http.HandlerFunc(forward.To(uri)))
	return r
}

// Head will register a pattern for HEAD requests.
func (r *Router) Head(path string) *Route {
	return r.add("HEAD", path, nil)
}

// Get will register a pattern for GET requests.
// It also registers pat for HEAD requests. If this needs to be overridden, use
// Head before Get with pat.
func (r *Router) Get(path string) *Route {
	return r.add("GET", path, nil)
}

// Post will register a pattern for POST requests.
func (r *Router) Post(path string) *Route {
	return r.add("POST", path, nil)
}

// Put will register a pattern for PUT requests.
func (r *Router) Put(path string) *Route {
	return r.add("PUT", path, nil)
}

// Delete will register a pattern for DELETE requests.
func (r *Router) Delete(path string) *Route {
	return r.add("DELETE", path, nil)
}

// Options will register a pattern for OPTIONS requests.
func (r *Router) Options(path string) *Route {
	return r.add("OPTIONS", path, nil)
}

// Patch will register a pattern for PATCH requests.
func (r *Router) Patch(path string) *Route {
	return r.add("PATCH", path, nil)
}

// All will register a pattern for any HTTP method.
func (r *Router) All(path string) *Route {
	return r.add("*", path, nil)
}

// Route will register a new route for the given pattern and HTTP method.
func (r *Router) Route(method, path string) *Route {
	return r.add(method, path, nil)
}

// add adds a new route to the router stack for the given method and path pattern.
func (r *Router) add(method, pat string, handler http.Handler) *Route {
	// Check route pattern is unique
	routes := r.Routes[method]
	for _, p1 := range routes {
		if p1.Pattern == pat {
			return p1
		}
	}

	route := NewRoute(pat)
	if handler != nil {
		route.Handler = handler
	}

	// Set middleware parent layer
	route.SetParent(r.Layer)

	// Register the route in the router stack
	r.Routes[method] = append(routes, route)

	n := len(pat)
	if n > 0 && pat[n-1] == '/' {
		r.add(method, pat[:n-1], route)
	}

	return route
}

// FindRoute tries to find a registered route who matches
// with the given method and path.
func (r *Router) FindRoute(method, path string) (*Route, error) {
	if _, route := r.match(method, path); route != nil {
		return route, nil
	}
	return nil, ErrNoRouteMatch
}

// match matches a registered route against the given method and URL path.
func (r *Router) match(method, path string) (url.Values, *Route) {
	// Match routes by given method
	if params, route := doMatch(r.Routes[method], path); route != nil {
		return params, route
	}
	// Match any method routes
	return doMatch(r.Routes["*"], path)
}

// doMatch is used to match a given path againts the registered routes pool.
func doMatch(routes []*Route, path string) (url.Values, *Route) {
	if routes == nil || len(routes) == 0 {
		return nil, nil
	}
	for _, route := range routes {
		if params, ok := route.Match(path); ok {
			return params, route
		}
	}
	return nil, nil
}

// Remove matches and removes the matched route from the router stack.
func (r *Router) Remove(method, path string) bool {
	routes := r.Routes[method]
	if routes == nil || len(routes) == 0 {
		return false
	}
	for i, route := range routes {
		if _, ok := route.Match(path); ok {
			r.Routes[method] = append(routes[:i], routes[i+1:]...)
			return true
		}
	}
	return false
}

// HandleHTTP matches r.URL.Path against its routing table
// using the rules described above.
func (r *Router) HandleHTTP(w http.ResponseWriter, req *http.Request, h http.Handler) {
	if params, route := r.match(req.Method, req.URL.Path); route != nil {
		if len(params) > 0 {
			req.URL.RawQuery = url.Values(params).Encode() + "&" + req.URL.RawQuery
		}
		r.Layer.Run(layer.RequestPhase, w, req, route)
		return
	}

	// If not found handler is defined, just continue.
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
		return
	}

	// If method not allowed behavior is enable,
	if r.ForceMethodNotAllowed && r.isMethodNotAllowed(w, req) {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	// If not found handler defines, use it
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
		return
	}

	h.ServeHTTP(w, req)
}

// isMethodNowAllowed is used to verify if the given route doesn't allows
// the given HTTP request method.
func (r *Router) isMethodNotAllowed(w http.ResponseWriter, req *http.Request) bool {
	allowed := make([]string, 0, len(r.Routes))
	for meth, handlers := range r.Routes {
		if meth == req.Method {
			continue
		}

		for _, r := range handlers {
			if _, ok := r.Match(req.URL.Path); ok {
				allowed = append(allowed, meth)
			}
		}
	}

	isAllowed := len(allowed) > 0
	if isAllowed {
		w.Header().Add("Allow", strings.Join(allowed, ", "))
	}
	return isAllowed
}

// Use attaches a new middleware handler for incoming HTTP traffic.
func (r *Router) Use(handler ...interface{}) *Router {
	r.Layer.Use(layer.RequestPhase, handler...)
	return r
}

// UsePhase attaches a new middleware handler to a specific phase.
func (r *Router) UsePhase(phase string, handler ...interface{}) *Router {
	r.Layer.Use(phase, handler...)
	return r
}

// UseFinalHandler uses a new middleware handler function as final handler.
func (r *Router) UseFinalHandler(fn http.Handler) *Router {
	r.Layer.UseFinalHandler(fn)
	return r
}

// SetParent sets a middleware parent layer.
// This method is tipically called via inversion of control from vinxi layer.
func (r *Router) SetParent(parent layer.Middleware) {
	r.Layer.SetParent(parent)
}

// Flush flushes the router middleware stack.
func (r *Router) Flush() {
	r.Layer.Flush()
}
