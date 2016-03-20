package forward

import (
	"crypto/tls"
	"gopkg.in/vinci-proxy/vinci.v0/utils"
	"io"
	"net"
	"net/http"
	"strings"
)

// websocketForwarder is a handler that can reverse proxy
// websocket traffic
type websocketForwarder struct {
	rewriter        ReqRewriter
	TLSClientConfig *tls.Config
}

// serveHTTP forwards websocket traffic
func (f *websocketForwarder) serveHTTP(w http.ResponseWriter, req *http.Request, ctx *handlerContext) {
	outReq := f.copyRequest(req)
	host := outReq.URL.Host
	dial := net.Dial

	// if host does not specify a port, use the default http port
	if !strings.Contains(host, ":") {
		if outReq.URL.Scheme == "wss" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	if outReq.URL.Scheme == "wss" {
		if f.TLSClientConfig == nil {
			f.TLSClientConfig = &tls.Config{}
		}
		dial = func(network, address string) (net.Conn, error) {
			return tls.Dial("tcp", host, f.TLSClientConfig)
		}
	}

	targetConn, err := dial("tcp", host)
	if err != nil {
		ctx.log.Errorf("Error dialing `%v`: %v", host, err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		ctx.log.Errorf("Unable to hijack the connection: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	underlyingConn, _, err := hijacker.Hijack()
	if err != nil {
		ctx.log.Errorf("Unable to hijack the connection: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	// it is now caller's responsibility to Close the underlying connection
	defer underlyingConn.Close()
	defer targetConn.Close()

	// write the modified incoming request to the dialed connection
	if err = outReq.Write(targetConn); err != nil {
		ctx.log.Errorf("Unable to copy request to target: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}
	errc := make(chan error, 2)
	replicate := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go replicate(targetConn, underlyingConn)
	go replicate(underlyingConn, targetConn)
	<-errc
}

// copyRequest makes a copy of the specified request.
func (f *websocketForwarder) copyRequest(req *http.Request) (outReq *http.Request) {
	outReq = new(http.Request)
	*outReq = *req
	outReq.URL = utils.CopyURL(req.URL)
	outReq.URL.Scheme = req.URL.Scheme
	outReq.URL.Host = req.URL.Host
	return outReq
}
