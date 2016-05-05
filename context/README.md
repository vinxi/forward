# context [![Build Status](https://travis-ci.org/vinxi/context.png)](https://travis-ci.org/vinxi/context) [![GoDoc](https://godoc.org/github.com/vinxi/vinxi/context?status.svg)](https://godoc.org/github.com/vinxi/vinxi/context) [![API](https://img.shields.io/badge/status-stable-green.svg?style=flat)](https://godoc.org/github.com/vinxi/vinxi/context) [![Coverage Status](https://coveralls.io/repos/github/vinxi/context/badge.svg?branch=master)](https://coveralls.io/github/vinxi/context?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/vinxi/context)](https://goreportcard.com/report/github.com/vinxi/context)

`context` package implements a simple, unobstructive context for request-aware data sharing across a middleware pipeline.

Originally based in [nbio/httpcontext](https://github.com/nbio/httpcontext).

## Installation

```bash
go get -u gopkg.in/vinxi/context.v0
```

## API

See [godoc](https://godoc.org/github.com/vinxi/context) reference.

## Example

```go
package main

import (
  "fmt"
  "gopkg.in/vinxi/context.v0"
  "gopkg.in/vinxi/vinxi.v0"
  "net/http"
)

func main() {
  fmt.Printf("Server listening on port: %d\n", 3100)
  vs := vinxi.NewServer(vinxi.ServerOptions{Host: "localhost", Port: 3100})

  vs.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
    context.Set(r, "foo", "bar")
    h.ServeHTTP(w, r)
  })

  vs.Use(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
    w.Header().Set("foo", context.GetString(r, "foo"))
    h.ServeHTTP(w, r)
  })

  vs.Forward("http://httpbin.org")

  err := vs.Listen()
  if err != nil {
    fmt.Errorf("Error: %s\n", err)
  }
}
```

## License

MIT
