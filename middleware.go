package forward

import (
	"net/http"
	"net/url"
)

// To returns an http.HandlerFunc that forwards the incoming request to
// the given URI server.
func To(uri string) func(w http.ResponseWriter, r *http.Request) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	fwd, err := New(PassHostHeader(true))
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = parsedURL.Scheme
		r.URL.Host = parsedURL.Host
		r.Host = parsedURL.Host

		// Forward the HTTP request
		fwd.ServeHTTP(w, r)
	}
}
