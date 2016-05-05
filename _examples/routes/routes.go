package main

import (
	"fmt"
	"net/http"

	"gopkg.in/vinxi/vinxi.v0"
)

func main() {
	vs := vinxi.NewServer(vinxi.ServerOptions{Port: 3100})
	v := vs.Vinxi

	v.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Server", "vinxi")
		h.ServeHTTP(w, r)
	})

	v.UsePhase("error", func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Server", "vinxi")
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	})

	v.Get("/ip").Forward("http://httpbin.org")
	v.Get("/headers").Forward("http://httpbin.org")
	v.Get("/image/:name").Forward("http://httpbin.org")
	v.All("/post").Forward("http://httpbin.org")

	v.Forward("http://example.com")

	fmt.Printf("Server listening on port: %d\n", 3100)
	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
