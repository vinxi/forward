package forward

const (
	// XForwardedProto stores the forwarder proto header key.
	XForwardedProto = "X-Forwarded-Proto"
	// XForwardedFor stores the forwar for header key.
	XForwardedFor = "X-Forwarded-For"
	// XForwardedHost stores the forward host header key.
	XForwardedHost = "X-Forwarded-Host"
	// XForwardedServer stores the forward server header key.
	XForwardedServer = "X-Forwarded-Server"
	// Connection stores the connection header key.
	Connection = "Connection"
	// KeepAlive stores the keep alive header key.
	KeepAlive = "Keep-Alive"
	// ProxyAuthenticate stores the proxy header key.
	ProxyAuthenticate = "Proxy-Authenticate"
	// ProxyAuthorization stores the proxy authorization header key.
	ProxyAuthorization = "Proxy-Authorization"
	// Te stores the proxy TE header key.
	Te = "Te" // canonicalized version of "TE"
	// Trailers stores the trailers header key.
	Trailers = "Trailers"
	// TransferEncoding stores the transfer encoding header key.
	TransferEncoding = "Transfer-Encoding"
	// Upgrade stores the update header key.
	Upgrade = "Upgrade"
	// ContentLength stores the content length header key.
	ContentLength = "Content-Length"
)

// HopHeaders stores the hop-by-hop headers.
// These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var HopHeaders = []string{
	Connection,
	KeepAlive,
	ProxyAuthenticate,
	ProxyAuthorization,
	Te, // canonicalized version of "TE"
	Trailers,
	TransferEncoding,
	Upgrade,
}
