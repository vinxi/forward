package sandbox

import (
	"net/http"
	"strconv"
	"time"
)

var (
	// DefaultPort stores the default TCP port to listen.
	DefaultPort = 8080

	// DefaultReadTimeout defines the maximum timeout for request read.
	DefaultReadTimeout = 60

	// DefaultWriteTimeout defines the maximum timeout for response write.
	DefaultWriteTimeout = 60
)

// ServerOptions represents the supported server options.
type ServerOptions struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
	Host         string
	CertFile     string
	KeyFile      string
}

// NewServer creates a new admin HTTP server.
func NewServer(o ServerOptions) *http.Server {
	// Apply default options
	if o.Port == 0 {
		o.Port = DefaultPort
	}
	if o.ReadTimeout == 0 {
		o.ReadTimeout = DefaultReadTimeout
	}
	if o.WriteTimeout == 0 {
		o.WriteTimeout = DefaultWriteTimeout
	}

	addr := o.Host + ":" + strconv.Itoa(o.Port)
	server := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    time.Duration(o.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.WriteTimeout) * time.Second,
	}

	return server
}

// Listen starts listening on the network.
func Listen(server *http.Server, opts ServerOptions) error {
	if opts.CertFile != "" && opts.KeyFile != "" {
		return server.ListenAndServeTLS(opts.CertFile, opts.KeyFile)
	}
	return server.ListenAndServe()
}
