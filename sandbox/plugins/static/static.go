package static

import (
	"net/http"

	"gopkg.in/vinxi/sandbox.v0"
)

// New creates a new static plugin who serves
// files of the given server local path.
func New(path string) sandbox.Plugin {
	return sandbox.NewPlugin("static", "serve static files", staticHandler(path))
}

func staticHandler(path string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.FileServer(http.Dir(path))
	}
}
