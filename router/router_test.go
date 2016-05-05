package router

import (
	"github.com/nbio/st"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

func TestRouterRemoveRoute(t *testing.T) {
	p := New()
	p.Get("/foo")
	p.All("/bar")

	st.Expect(t, len(p.Routes["GET"]), 1)
	st.Expect(t, len(p.Routes["*"]), 1)

	st.Expect(t, routeExists(p, "*", "/bar"), true)
	st.Expect(t, routeExists(p, "GET", "/foo"), true)

	st.Expect(t, p.Remove("GET", "/foo"), true)
	st.Expect(t, len(p.Routes["GET"]), 0)
	st.Expect(t, routeExists(p, "GET", "/foo"), false)

	st.Expect(t, p.Remove("*", "/bar"), true)
	st.Expect(t, len(p.Routes["*"]), 0)
	st.Expect(t, routeExists(p, "*", "/bar"), false)

	st.Expect(t, p.Remove("*", "/baz"), false)
}

func TestRoutingHit(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/:name").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		st.Expect(t, r.URL.Query().Get(":name"), "keith")
	}))

	p.HandleHTTP(nil, newRequest("GET", "/foo/keith?a=b", nil), nil)
	if !ok {
		t.Error("handler not called")
	}
}

func TestRoutingMethodNotAllowed(t *testing.T) {
	p := New()
	p.ForceMethodNotAllowed = true

	var ok bool
	p.Post("/foo/:name").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	p.Put("/foo/:name").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	r := httptest.NewRecorder()
	var final bool
	p.HandleHTTP(r, newRequest("GET", "/foo/keith", nil), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		final = true
	}))

	st.Expect(t, ok, false)
	st.Expect(t, final, false)
	st.Expect(t, r.Code, http.StatusMethodNotAllowed)

	got := strings.Split(r.Header().Get("Allow"), ", ")
	sort.Strings(got)
	want := []string{"POST", "PUT"}
	st.Expect(t, got, want)
}

// Check to make sure we don't pollute the Raw Query when we have no parameters
func TestNoParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		st.Expect(t, r.URL.RawQuery, "")
	}))

	p.HandleHTTP(nil, newRequest("GET", "/foo/", nil), nil)
	st.Expect(t, ok, true)
}

// Check to make sure we don't pollute the Raw Query when there are parameters but no pattern variables
func TestOnlyUserParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		st.Expect(t, r.URL.RawQuery, "a=b")
	}))

	p.HandleHTTP(nil, newRequest("GET", "/foo/?a=b", nil), nil)
	st.Expect(t, ok, true)
}

func TestImplicitRedirect(t *testing.T) {
	p := New()
	p.Get("/foo/").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	res := httptest.NewRecorder()
	p.HandleHTTP(res, newRequest("GET", "/foo", nil), nil)
	st.Expect(t, res.Code, 200)
	st.Expect(t, res.Header().Get("Location"), "")

	p = New()
	p.Get("/foo").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	p.Get("/foo/").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	res = httptest.NewRecorder()
	p.HandleHTTP(res, newRequest("GET", "/foo", nil), nil)
	st.Expect(t, res.Code, 200)

	p = New()
	p.Get("/hello/:name/").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	res = httptest.NewRecorder()
	p.HandleHTTP(res, newRequest("GET", "/hello/bob?a=b#f", nil), nil)
	st.Expect(t, res.Code, 200)
}

func TestNotFound(t *testing.T) {
	p := New()
	p.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(123)
	})
	p.Post("/bar").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	for _, path := range []string{"/foo", "/bar"} {
		res := httptest.NewRecorder()
		p.HandleHTTP(res, newRequest("GET", path, nil), nil)
		st.Expect(t, res.Code, 123)
	}
}

func TestMethodPatch(t *testing.T) {
	p := New()
	p.Patch("/foo/bar").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Test to see if we get a 405 Method Not Allowed errors from trying to
	// issue a GET request to a handler that only supports the PATCH method.
	res := httptest.NewRecorder()
	res.Code = http.StatusMethodNotAllowed
	var final bool
	p.HandleHTTP(res, newRequest("GET", "/foo/bar", nil), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		final = true
	}))

	st.Expect(t, final, true)
	st.Expect(t, res.Code, http.StatusMethodNotAllowed)

	// Now, test to see if we get a 200 OK from issuing a PATCH request to
	// the same handler.
	res = httptest.NewRecorder()
	p.HandleHTTP(res, newRequest("PATCH", "/foo/bar", nil), nil)
	st.Expect(t, res.Code, http.StatusOK)
}

func BenchmarkPatternMatching(b *testing.B) {
	p := New()
	p.Get("/hello/:name").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		r := newRequest("GET", "/hello/blake", nil)
		b.StartTimer()
		p.HandleHTTP(nil, r, nil)
	}
}

func newRequest(method, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err)
	}
	return req
}

func routeExists(r *Router, method, path string) bool {
	route, _ := r.FindRoute(method, path)
	return route != nil
}
