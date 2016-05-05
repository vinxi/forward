package layer

import (
	"github.com/nbio/st"
	"gopkg.in/vinxi/utils.v0"
	"net/http"
	"testing"
)

type vinxiHandler struct{}

func (vh vinxiHandler) HandleHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	w.Header().Set("foo", "bar")
	h.ServeHTTP(w, r)
}

func TestAdaptMiddlewareFunc(t *testing.T) {
	middlewareFunc := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("foo", "bar")
			h.ServeHTTP(w, r)
		})
	}

	w := utils.NewWriterStub()
	req := &http.Request{}

	adaptedFunc := AdaptFunc(middlewareFunc)
	adaptedFunc(FinalHandler).ServeHTTP(w, req)

	st.Expect(t, w.Header().Get("foo"), "bar")
	st.Expect(t, w.Code, 502)
}

func TestAdaptNegroniInterface(t *testing.T) {
	middlewareFunc := func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("foo", "bar")
		h.ServeHTTP(w, r)
	}

	w := utils.NewWriterStub()
	req := &http.Request{}

	adaptedFunc := AdaptFunc(middlewareFunc)
	adaptedFunc(FinalHandler).ServeHTTP(w, req)

	st.Expect(t, w.Header().Get("foo"), "bar")
	st.Expect(t, w.Code, 502)
}

func TestStandardHttpHandlerInterface(t *testing.T) {
	middlewareFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", "bar")
	}

	w := utils.NewWriterStub()
	req := &http.Request{}

	adaptedFunc := AdaptFunc(middlewareFunc)
	adaptedFunc(FinalHandler).ServeHTTP(w, req)

	st.Expect(t, w.Header().Get("foo"), "bar")
	st.Reject(t, w.Code, 502)
}

func TestStandardHttpHandler(t *testing.T) {
	middlewareFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", "bar")
	})

	w := utils.NewWriterStub()
	req := &http.Request{}

	adaptedFunc := AdaptFunc(middlewareFunc)
	adaptedFunc(FinalHandler).ServeHTTP(w, req)

	st.Expect(t, w.Header().Get("foo"), "bar")
	st.Reject(t, w.Code, 502)
}

func TestVinciHandler(t *testing.T) {
	middlewareFunc := vinxiHandler{}

	w := utils.NewWriterStub()
	req := &http.Request{}

	adaptedFunc := AdaptFunc(middlewareFunc)
	adaptedFunc(FinalHandler).ServeHTTP(w, req)

	st.Expect(t, w.Header().Get("foo"), "bar")
	st.Expect(t, w.Code, 502)
}
