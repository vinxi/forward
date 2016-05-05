#### Simple forward proxy

Forwards all the traffic to a specific host.

```go
package main

import (
  "fmt"
  "gopkg.in/vinxi/vinxi.v0"
)

func main() {
  vs := vinxi.NewServer(vinxi.ServerOptions{Port: 3100})

  // Forward all the traffic to httpbin.org
  vs.Forward("http://httpbin.org")

  fmt.Printf("Server listening on port: %d\n", 3100)
  err := vs.Listen()
  if err != nil {
    fmt.Errorf("Error: %s\n", err)
  }
}
```

#### Multiplexer composition

Uses the built-in multiplexer for traffic handling composition.

```go
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
```

#### Simple traffic routing

Uses the built-in router for handle and forward traffic.

```go
package main

import (
  "fmt"
  "net/http"
  "gopkg.in/vinxi/vinxi.v0"
)

func main() {
  vs := vinxi.NewServer(vinxi.ServerOptions{Host: "localhost", Port: 3100})
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
```
