package layer

import "net/http"

// Handler represents the vinxi specific supported interface
// that can be implemented by middleware handlers.
type Handler interface {
	HandleHTTP(http.ResponseWriter, *http.Request, http.Handler)
}

// PartialHandler represents the vinxi specific supported interface
// that applies partial function application and can be implemented
// by middleware handlers.
type PartialHandler interface {
	HandleHTTP(http.Handler) func(http.ResponseWriter, *http.Request)
}

// HandlerFunc represents the required function interface for simple middleware handlers.
type HandlerFunc func(http.ResponseWriter, *http.Request)

// HandlerFuncNext represents a Negroni-like handler function notation.
type HandlerFuncNext func(http.ResponseWriter, *http.Request, http.Handler)

// MiddlewareFunc represents the http.Handler -> http.Handler capable interface.
type MiddlewareFunc func(http.Handler) http.Handler

// MiddlewareHandlerFunc represents the http.Handler -> http.HandlerFunc capable interface.
type MiddlewareHandlerFunc func(http.Handler) func(http.ResponseWriter, *http.Request)

// Registrable represents the required interface implemented by middleware capable handlers
// to register one or multiple middleware phases.
//
// This is mostly used as inversion of control mecanish allowing to third-party middleware
// implementors the ability to register multiple middleware handlers transparently.
//
// For instance, you can register request and error handlers:
//
//   func (s *MyStruct) Register(mw layer.Middleware) {
//      mw.Use("request", s.requestHandler)
//      mw.Use("error", s.errorHandler)
//   }
//
type Registrable interface {
	// Register is designed to allow the plugin developers
	// to attach multiple middleware layers passing the current middleware layer.
	Register(Middleware)
}

// AdaptFunc adapts the given function polumorphic interface
// casting into a MiddlewareFunc capable interface.
//
// Currently support five different interface notations,
// wrapping it accordingly to make homogeneus.
func AdaptFunc(h interface{}) MiddlewareFunc {
	// Vinxi/Alice interface
	if mw, ok := h.(func(h http.Handler) http.Handler); ok {
		return MiddlewareFunc(mw)
	}

	// http.Handler -> http.HandlerFunc interface
	if mw, ok := h.(func(http.Handler) func(http.ResponseWriter, *http.Request)); ok {
		return adaptMiddlewareHandlerFunc(mw)
	}

	// Negroni like interface
	if mw, ok := h.(func(w http.ResponseWriter, r *http.Request, h http.Handler)); ok {
		return adaptHandlerFuncNext(mw)
	}

	// Standard net/http function handler interface
	if mw, ok := h.(func(http.ResponseWriter, *http.Request)); ok {
		return adaptHandlerFunc(mw)
	}

	// Standard net/http handler
	if mw, ok := h.(http.Handler); ok {
		return adaptNativeHandler(mw)
	}

	// Vinxi's built-in handler interface
	if mw, ok := h.(Handler); ok {
		return adaptHandler(mw)
	}

	// Vinxi's built-in partial handler interface
	if mw, ok := h.(PartialHandler); ok {
		return adaptPartialHandler(mw)
	}

	return nil
}

func adaptHandlerFunc(fn HandlerFunc) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(fn)
	}
}

func adaptMiddlewareHandlerFunc(fn MiddlewareHandlerFunc) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(fn(h))
	}
}

func adaptHandlerFuncNext(fn HandlerFuncNext) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r, h)
		})
	}
}

func adaptHandler(fn Handler) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn.HandleHTTP(w, r, h)
		})
	}
}

func adaptPartialHandler(fn PartialHandler) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		next := fn.HandleHTTP(h)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		})
	}
}

func adaptNativeHandler(fn http.Handler) MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return fn
	}
}
