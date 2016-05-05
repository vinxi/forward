# router [![Build Status](https://travis-ci.org/vinxi/router.png)](https://travis-ci.org/vinxi/router) [![GoDoc](https://godoc.org/github.com/vinxi/router?status.svg)](https://godoc.org/github.com/vinxi/router) [![Coverage Status](https://coveralls.io/repos/github/vinxi/router/badge.svg?branch=master)](https://coveralls.io/github/vinxi/router?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/vinxi/router)](https://goreportcard.com/report/github.com/vinxi/router) [![API](https://img.shields.io/badge/vinxi-core-green.svg?style=flat)](https://godoc.org/github.com/vinxi/layer) 

Featured and fast router used by vinxi. Provides a hierarchical middleware layer.

Originally based in [bmizerany/pat](https://github.com/bmizerany/pat).

## Installation

```bash
go get -u gopkg.in/vinxi/router.v0
```

## API

See [godoc reference](https://godoc.org/github.com/vinxi/router) for detailed API documentation.

## Examples

#### Router 

```go
package main

import (
  "fmt"
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

  vs.Use(r)
  vs.Forward("http://example.com")

  err := vs.Listen()
  if err != nil {
    fmt.Errorf("Error: %s\n", err)
  }
}
```

#### Virtual host like muxer router 

```go
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
```

## License

[MIT](https://opensource.org/licenses/MIT).
