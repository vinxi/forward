package main

import (
	"fmt"
	"net/http"

	"gopkg.in/vinxi/vinxi.v0"
)

const port = 3100

func main() {
	vs := vinxi.NewServer(vinxi.ServerOptions{Port: port})
	v := vs.Vinxi

	// Register a simple middleware handler
	v.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		w.Header().Set("Server", "vinxi")
		h.ServeHTTP(w, r)
	})

	// Compose the API gateway
	v.All("/users/").Forward("http://127.0.0.1:3200")
	v.All("/posts/").Forward("http://127.0.0.1:3201")
	v.All("/files/").Forward("http://127.0.0.1:3202")
	v.All("/authentication/").Forward("http://127.0.0.1:3203")

	// Default forward to traffic to httpbin.org
	v.Forward("http://httpbin.org")

	fmt.Printf("Server listening on port: %d\n", port)
	err := vs.Listen()
	if err != nil {
		fmt.Errorf("Error: %s\n", err)
	}
}
