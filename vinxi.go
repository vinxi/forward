package vinxi

import (
	"net/http"

	"gopkg.in/vinxi/context.v0"
	"gopkg.in/vinxi/forward.v0"
	"gopkg.in/vinxi/layer.v0"
	"gopkg.in/vinxi/mux.v0"
	"gopkg.in/vinxi/router.v0"
)

// DefaultForwarder stores the default http.Handler to be used to forward the traffic.
// By default the proxy will reply with 502 Bad Gateway if no custom forwarder is defined.
var DefaultForwarder, _ = forward.New(forward.PassHostHeader(true))

// Middleware defines the required interface implemented
// by public middleware capable entities in the vinxi ecosystem.
type Middleware interface {
	// Use is used to register one or multiple middleware handlers.
	Use(...interface{}) Middleware
	// UsePhase is used to register one or multiple middleware
	// handlers for a specific middleware phase.
	UsePhase(string, ...interface{}) Middleware
	// UseFinalHandler is used to register the final request handler
	// usually to define the error or forward handlers.
	UseFinalHandler(http.Handler) Middleware
	// SetParent allows hierarchical middleware inheritance.
	SetParent(layer.Middleware)
}

// Vinxi represents the vinxi proxy layer.
type Vinxi struct {
	// Layer stores the proxy level middleware layer.
	Layer *layer.Layer
	// Router stores the built-in router.
	Router *router.Router
}

// New creates a new vinxi proxy layer with default fields.
func New() *Vinxi {
	v := &Vinxi{Layer: layer.New(), Router: router.New()}
	// Bind router with parent layer
	v.Router.SetParent(v.Layer)
	// Register the router in the middleware tail (this should change in the future)
	v.Layer.UsePriority("request", layer.Tail, v.Router)
	// Use the default forwarder as final middleware handler
	v.UseFinalHandler(DefaultForwarder)
	return v
}

// Get will register a pattern for GET requests.
// It also registers pat for HEAD requests. If this needs to be overridden, use
// Head before Get with pat.
func (v *Vinxi) Get(path string) *router.Route {
	return v.Route("GET", path)
}

// Post will register a pattern for POST requests.
func (v *Vinxi) Post(path string) *router.Route {
	return v.Route("POST", path)
}

// Put will register a pattern for PUT requests.
func (v *Vinxi) Put(path string) *router.Route {
	return v.Route("PUT", path)
}

// Delete will register a pattern for DELETE requests.
func (v *Vinxi) Delete(path string) *router.Route {
	return v.Route("DELETE", path)
}

// Options will register a pattern for OPTIONS requests.
func (v *Vinxi) Options(path string) *router.Route {
	return v.Route("OPTIONS", path)
}

// Patch will register a pattern for PATCH requests.
func (v *Vinxi) Patch(path string) *router.Route {
	return v.Route("PATCH", path)
}

// All will register a pattern for any HTTP method.
func (v *Vinxi) All(path string) *router.Route {
	return v.Route("*", path)
}

// Route will register a new route for the given pattern and HTTP method.
func (v *Vinxi) Route(method, path string) *router.Route {
	return v.Router.Route(method, path)
}

// Forward defines the default URL to forward incoming traffic.
func (v *Vinxi) Forward(uri string) *Vinxi {
	return v.UseFinalHandler(http.HandlerFunc(forward.To(uri)))
}

// Use attaches a new middleware handler for incoming HTTP traffic.
func (v *Vinxi) Use(handler ...interface{}) *Vinxi {
	v.Layer.Use(layer.RequestPhase, handler...)
	return v
}

// UsePhase attaches a new middleware handler to a specific phase.
func (v *Vinxi) UsePhase(phase string, handler ...interface{}) *Vinxi {
	v.Layer.Use(phase, handler...)
	return v
}

// UseFinalHandler uses a new middleware handler function as final handler.
func (v *Vinxi) UseFinalHandler(fn http.Handler) *Vinxi {
	v.Layer.UseFinalHandler(fn)
	return v
}

// SetForwader sets the default final traffic forwarder.
func (v *Vinxi) SetForwader(fn http.Handler) *Vinxi {
	v.Layer.UseFinalHandler(fn)
	return v
}

// Flush flushes all the middleware stack.
func (v *Vinxi) Flush() {
	v.Layer.Flush()
}

// Mux creates a new multiplexer based on the given matcher functions.
func (v *Vinxi) Mux(matchers ...mux.Matcher) *mux.Mux {
	m := mux.New()
	// Register matchers
	m.If(matchers...)
	// Bind multiplexer to the current vinxi layer
	v.Use(m)
	return m
}

// SetParent sets a parent middleware layer.
func (v *Vinxi) SetParent(parent layer.Middleware) {
	v.Layer.SetParent(parent)
}

// NewServer creates a new http.Server.
func (v *Vinxi) NewServer(opts ServerOptions) *Server {
	srv := NewServer(opts)
	v.BindServer(srv.Server)
	return srv
}

// ServeAndListen creates a new http.Server and starts listening on the network.
func (v *Vinxi) ServeAndListen(opts ServerOptions) (*Server, error) {
	srv := NewServer(opts)
	v.BindServer(srv.Server)
	return srv, srv.Listen()
}

// BindServer binds the vinxi HTTP handler to the given http.Server.
func (v *Vinxi) BindServer(server *http.Server) {
	server.Handler = v
}

// ServeHTTP implements the required http.Handler interface to handle incoming traffic.
func (v *Vinxi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Expose original request host
	context.Set(r, "vinxi.host", r.Host)
	// Define target URL
	r.URL.Host = r.Host
	// Run the incoming request middleware layer
	v.Layer.Run("request", w, r, nil)
}
