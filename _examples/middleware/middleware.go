package main

import (
	"fmt"
	"net/http"

	"gopkg.in/vinxi/mux.v0"
	"gopkg.in/vinxi/vinxi.v0"
)

const port = 3100

func main() {
	vs := vinxi.NewServer(vinxi.ServerOptions{Port: port})
	v := vs.Vinxi

	// vinxi level middleware handler
	v.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Server", "vinxi")
		h.ServeHTTP(w, r)
	})

	// Middleware via multiplexer (only for GET requests)
	v.Mux(mux.MatchMethod("GET")).Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Operation", "read")
		h.ServeHTTP(w, r)
	})

	// Default target to forward all the traffic
	v.Forward("http://httpbin.org")

	fmt.Printf("Server listening on port: %d\n", port)
	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
