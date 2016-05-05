package main

import (
	"fmt"
	"gopkg.in/vinxi/mux.v0"
	"gopkg.in/vinxi/router.v0"
	"gopkg.in/vinxi/vinxi.v0"
	"net/http"
)

func main() {
	fmt.Printf("Server listening on port: %d\n", 3100)
	vs := vinxi.NewServer(vinxi.ServerOptions{Host: "localhost", Port: 3100})

	r := router.New()
	r.Get("/get").Forward("http://httpbin.org")
	r.Get("/headers").Forward("http://httpbin.org")
	r.Get("/image/:format").Forward("http://httpbin.org")
	r.Get("/say").Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, foo"))
	}))

	// Create a host header multiplexer
	muxer := mux.Host("localhost:3100")
	muxer.Use(r)

	vs.Use(muxer)
	vs.Forward("http://example.com")

	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
