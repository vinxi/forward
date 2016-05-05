package main

import (
	"fmt"
	"net/http"

	"gopkg.in/vinxi/forward.v0"
	"gopkg.in/vinxi/mux.v0"
	"gopkg.in/vinxi/router.v0"
	"gopkg.in/vinxi/vinxi.v0"
)

func main() {
	fmt.Printf("Server listening on port: %d\n", 3100)
	vs := vinxi.NewServer(vinxi.ServerOptions{Host: "localhost", Port: 3100})

	m := mux.New()
	m.If(mux.MatchHost("localhost:3100"))

	m.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Server", "vinxi")
		h.ServeHTTP(w, r)
	})

	m.Use(router.NewRoute("/foo").Handler(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("foo bar"))
	}))

	m.Use(forward.To("http://127.0.0.1:8080"))

	vs.Vinci.Use(m)
	vs.Vinci.Forward("http://127.0.0.1")

	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
