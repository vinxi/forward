package forward

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/vinxi/utils.v0"
)

// httpForwarder is a handler that can reverse proxy
// HTTP traffic
type httpForwarder struct {
	roundTripper http.RoundTripper
	rewriter     ReqRewriter
	passHost     bool
}

// serveHTTP forwards HTTP traffic using the configured transport
func (f *httpForwarder) serveHTTP(w http.ResponseWriter, req *http.Request, ctx *handlerContext) {
	start := time.Now().UTC()
	response, err := f.roundTripper.RoundTrip(f.copyRequest(req, req.URL))
	if err != nil {
		ctx.log.Errorf("Error forwarding to %v, err: %v", req.URL, err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}

	if req.TLS != nil {
		ctx.log.Infof("Round trip: %v, code: %v, duration: %v tls:version: %x, tls:resume:%t, tls:csuite:%x, tls:server:%v",
			req.URL, response.StatusCode, time.Now().UTC().Sub(start),
			req.TLS.Version,
			req.TLS.DidResume,
			req.TLS.CipherSuite,
			req.TLS.ServerName)
	} else {
		ctx.log.Infof("Round trip: %v, code: %v, duration: %v",
			req.URL, response.StatusCode, time.Now().UTC().Sub(start))
	}

	utils.CopyHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	written, err := io.Copy(w, response.Body)
	defer response.Body.Close()

	if err != nil {
		ctx.log.Errorf("Error copying upstream response Body: %v", err)
		ctx.errHandler.ServeHTTP(w, req, err)
		return
	}

	if written != 0 {
		w.Header().Set(ContentLength, strconv.FormatInt(written, 10))
	}
}

// copyRequest makes a copy of the specified request to be sent using the configured
// transport
func (f *httpForwarder) copyRequest(req *http.Request, u *url.URL) *http.Request {
	outReq := new(http.Request)
	*outReq = *req // includes shallow copies of maps, but we handle this below

	outReq.URL = utils.CopyURL(req.URL)
	outReq.URL.Host = u.Host
	outReq.URL.Opaque = req.RequestURI
	outReq.URL.Scheme = u.Scheme
	if outReq.URL.Scheme == "" {
		outReq.URL.Scheme = "http"
	}
	// raw query is already included in RequestURI, so ignore it to avoid dupes
	outReq.URL.RawQuery = ""
	// Do not pass client Host header unless optsetter PassHostHeader is set.
	if !f.passHost {
		outReq.Host = u.Host
	}
	outReq.Proto = "HTTP/1.1"
	outReq.ProtoMajor = 1
	outReq.ProtoMinor = 1

	// Overwrite close flag so we can keep persistent connection for the backend servers
	outReq.Close = false

	outReq.Header = make(http.Header)
	utils.CopyHeaders(outReq.Header, req.Header)

	if f.rewriter != nil {
		f.rewriter.Rewrite(outReq)
	}
	return outReq
}
