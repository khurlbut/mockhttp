package mockhttp

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	netURL "net/url"
	"strings"
)

type HTTPFake struct {
	Server          *httptest.Server
	RequestHandlers []*Request
}

func Server() *HTTPFake {
	server := &HTTPFake{
		RequestHandlers: []*Request{},
	}

	server.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rh := server.findHandler(r)
		if rh == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if rh.CustomHandle != nil {
			rh.CustomHandle(w, r, rh)
			return
		}
		DefaultResponder(w, r, rh)
	}))

	return server
}

func (f *HTTPFake) Start(ip string, port string) *HTTPFake {
	fmt.Println("starting server")
	f.Server.Listener = myLocalListener(ip, port)
	f.Server.Start()
	return f
}

func (f *HTTPFake) Close() {
	fmt.Println("closing server...")
	f.Server.Close()
}

func myLocalListener(ip string, port string) net.Listener {
	l, err := net.Listen("tcp", ip+":"+port)
	// l, err := net.Listen("tcp", "172.17.0.2:8181")
	if err != nil {
		fmt.Println("--- TCP FAILED! Using TCP6! ---")
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func (f *HTTPFake) NewHandler() *Request {
	rh := NewRequest()
	f.RequestHandlers = append(f.RequestHandlers, rh)
	return rh
}

func (f *HTTPFake) ResolveURL(path string, args ...interface{}) string {
	format := f.Server.URL + path
	return fmt.Sprintf(format, args...)
}

func (f *HTTPFake) Reset() *HTTPFake {
	f.RequestHandlers = []*Request{}
	return f
}

func (f *HTTPFake) findHandler(r *http.Request) *Request {
	founds := []*Request{}
	url := r.URL.String()
	path := getURLPath(url)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	for _, rh := range f.RequestHandlers {
		if rh.Method != r.Method {
			continue
		}

		rhURL, _ := netURL.QueryUnescape(rh.URL.String())

		if rhURL == url {
			return rh
		}

		if getURLPath(rhURL) == path {
			founds = append(founds, rh)
		}
	}

	if len(founds) == 1 {
		return founds[0]
	}

	return nil
}

func getURLPath(url string) string {
	return strings.Split(url, "?")[0]
}