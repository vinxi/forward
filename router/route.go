package router

import (
	"net/http"
	"net/url"

	"gopkg.in/vinxi/forward.v0"
	"gopkg.in/vinxi/layer.v0"
)

// DefaultForwarder stores the default http.Handler to be used to forward the traffic.
// By default the proxy will reply with 502 Bad Gateway if no custom forwarder is defined.
var DefaultForwarder = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fwd, _ := forward.New(forward.PassHostHeader(true))
	fwd.ServeHTTP(w, r)
})

// Route represents an HTTP route based on the given
type Route struct {
	// Pattern stores the path pattern
	Pattern string

	// Layer provides the middleware layer at route level.
	Layer *layer.Layer

	// Handler stores the final route handler function.
	Handler http.Handler
}

// NewRoute creates a new Route for the given URL path pattern.
func NewRoute(pattern string) *Route {
	return &Route{Pattern: pattern, Layer: layer.New()}
}

// Handle defines a custom route final handler function.
// Use this method only if you really want to handle the
// route in a very specific way.
func (r *Route) Handle(handler http.HandlerFunc) {
	r.Handler = http.HandlerFunc(handler)
}

// Match matches an incoming request againts the registered matchers
// returning true if all matchers passes.
func (r *Route) Match(path string) (url.Values, bool) {
	return Match(r.Pattern, path)
}

// Forward defines the default URL to forward incoming traffic.
func (r *Route) Forward(uri string) {
	r.Layer.UseFinalHandler(http.HandlerFunc(forward.To(uri)))
}

// Use attaches a new middleware handler for incoming HTTP traffic.
func (r *Route) Use(handler interface{}) *Route {
	if r.Handler == nil {
		r.Handler = DefaultForwarder
	}
	r.Layer.Use(layer.RequestPhase, handler)
	return r
}

// UsePhase attaches a new middleware handler to a specific phase.
func (r *Route) UsePhase(phase string, handler interface{}) *Route {
	r.Layer.Use(phase, handler)
	return r
}

// UseFinalHandler attaches a new http.Handler as final middleware layer handler.
func (r *Route) UseFinalHandler(handler http.Handler) *Route {
	r.Layer.UseFinalHandler(handler)
	return r
}

// SetParent sets a middleware parent layer.
// This method is tipically called via inversion of control from parent router.
func (r *Route) SetParent(parent layer.Middleware) {
	r.Layer.SetParent(parent)
}

// Flush flushes the route level middleware stack.
func (r *Route) Flush() {
	r.Layer.Flush()
}

// ServeHTTP handlers the incoming request and implemented the vinxi specific handler interface.
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Layer.Run(layer.RequestPhase, w, req, r.Handler)
}
