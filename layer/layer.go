// Package layer implements a simple HTTP server middleware layer
// used internally by vinxi to compose and trigger the middleware chain.
package layer

import (
	"net/http"

	"gopkg.in/vinxi/context.v0"
)

const (
	// ErrorPhase defines error middleware phase idenfitier.
	ErrorPhase = "error"
	// RequestPhase defines the default middleware phase for request.
	RequestPhase = "request"
)

// FinalHandler stores the default http.Handler used as final middleware chain.
// You can customize this handler in order to reply with a default error response.
var FinalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(502)
	w.Write([]byte("Bad Gateway"))
})

// FinalErrorHandler stores the default http.Handler used as final middleware chain.
// You can customize this handler in order to reply with a default error response.
var FinalErrorHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	w.Write([]byte("Proxy Error"))
})

// Runnable represents the required interface for a runnable
type Runnable interface {
	Run(string, http.ResponseWriter, *http.Request, http.Handler)
}

// Pluggable represents a middleware pluggable interface implementing
// the required methods to plug in middleware handlers.
type Pluggable interface {
	// Use method is used to register a new middleware handler in the stack.
	Use(phase string, handler ...interface{})
	// UsePriority method is used to register a new middleware handler in a specific phase.
	UsePriority(string, Priority, ...interface{})
	// UseFinalHandler defines the middleware handler terminator
	UseFinalHandler(handler http.Handler)
	// SetParent allows hierarchical middleware inheritance.
	SetParent(Middleware)
}

// Middleware especifies the required interface that must be
// implemented by third-party middleware capable interfaces.
type Middleware interface {
	// Middleware is also a Runnable interface.
	Runnable
	// Middleware is also a Pluggable interface.
	Pluggable
	// Flush flushes the middleware handlers pool.
	Flush()
}

// Pool represents the phase-specific stack to store middleware functions.
type Pool map[string]*Stack

// Layer type represent an HTTP domain
// specific middleware layer with hieritance support.
type Layer struct {
	// finalHandler stores the final middleware chain handler.
	finalHandler http.Handler
	// parent stores the parent middleware layer to use. Use SetParent(parent).
	parent Middleware
	// Pool stores the phase-specific middleware handlers stack.
	Pool Pool
}

// New creates a new middleware layer.
func New() *Layer {
	return &Layer{Pool: make(Pool), finalHandler: FinalHandler}
}

// Flush flushes the middleware pool.
func (s *Layer) Flush() {
	s.Pool = make(Pool)
}

// Use registers new handlers for the given phase in the middleware stack.
func (s *Layer) Use(phase string, handler ...interface{}) {
	s.use(phase, Normal, handler...)
}

// UsePriority registers new handlers for the given phase in the middleware stack with a custom priority.
func (s *Layer) UsePriority(phase string, priority Priority, handler ...interface{}) {
	s.use(phase, priority, handler...)
}

// UseFinalHandler defines an http.Handler as final middleware call chain handler.
// This handler is tipically responsible of replying with a custom response
// or error (e.g: cannot route the request).
func (s *Layer) UseFinalHandler(fn http.Handler) {
	s.finalHandler = fn
}

// SetParent sets a new middleware layer as parent layer,
// allowing to trigger ancestors layer from the current one.
func (s *Layer) SetParent(parent Middleware) {
	s.parent = parent
}

// use is used internally to register one or multiple middleware handlers
// in the middleware pool in the given phase and ordered by the given priority.
func (s *Layer) use(phase string, priority Priority, handler ...interface{}) *Layer {
	if s.Pool[phase] == nil {
		s.Pool[phase] = &Stack{}
	}

	stack := s.Pool[phase]
	for _, h := range handler {
		register(s, stack, priority, h)
	}

	return s
}

// register infers the handler interface and registers it in the given middleware stack.
func register(layer *Layer, stack *Stack, priority Priority, handler interface{}) {
	// Vinci's registrable interface
	if r, ok := handler.(Registrable); ok {
		r.Register(layer)
		return
	}

	// Otherwise infer the function interface
	mw := AdaptFunc(handler)
	if mw == nil {
		panic("vinxi: unsupported middleware interface")
	}

	stack.Push(priority, mw)
}

// Run triggers the middleware call chain for the given phase.
// In case of panic, it will be recovered transparently and trigger the error middleware chain.ç
func (s *Layer) Run(phase string, w http.ResponseWriter, r *http.Request, h http.Handler) {
	// In case of panic we want to handle it accordingly
	defer func() {
		if phase == "error" {
			return
		}
		if re := recover(); re != nil {
			s.runRecoverError(re, w, r)
		}
	}()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.run(phase, w, r, h)
	})

	// Run parent layer for the given phase, if present
	if phase != RequestPhase && s.parent != nil {
		s.parent.Run(phase, w, r, next)
		return
	}

	// Otherwise run the current layer
	next.ServeHTTP(w, r)
}

// run runs the current layer middleware chain for the given phase.
func (s *Layer) run(phase string, w http.ResponseWriter, r *http.Request, h http.Handler) {
	// Use default final handler if no one is passed
	if h == nil {
		h = s.finalHandler
	}

	// Get registered middleware handlers for the current phase
	stack, ok := s.Pool[phase]
	if !ok {
		h.ServeHTTP(w, r)
		return
	}

	// Build the middleware handlers call chain
	queue := stack.Join()
	for i := len(queue) - 1; i >= 0; i-- {
		h = queue[i](h)
	}

	// Trigger the first middleware handler
	h.ServeHTTP(w, r)
}

// runRecoverError runs the current layer error phase middleware chain
// triggering the parent layer if necessary.
func (s *Layer) runRecoverError(rerr interface{}, w http.ResponseWriter, r *http.Request) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If no parent, run default error final handler
		if s.parent == nil {
			FinalErrorHandler.ServeHTTP(w, r)
			return
		}
		// If parent layer exists, trigger it
		s.parent.Run("error", w, r, FinalErrorHandler)
	})

	// Expose error via context. This may change in a future.
	context.Set(r, "vinxi.error", rerr)
	s.run("error", w, r, next)
}
