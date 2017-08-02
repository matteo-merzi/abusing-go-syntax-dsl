package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type RequestFilter func(*http.Request) bool

func CIDR(cidrs ...string) RequestFilter {
	nets := make([]*net.IPNet, len(cidrs))
	for i, cidr := range cidrs {
		// TODO: handle err
		_, nets[i], _ = net.ParseCIDR(cidr)
	}
	return func(r *http.Request) bool {
		// TODO: handle err
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		for _, netw := range nets {
			if netw.Contains(ip) {
				return true
			}
		}
		return false
	}
}

func PasswordHeader(password string) RequestFilter {
	return func(r *http.Request) bool {
		return r.Header.Get("X-Password") == password
	}
}

func Method(methods ...string) RequestFilter {
	return func(r *http.Request) bool {
		for _, m := range methods {
			if r.Method == m {
				return true
			}
		}
		return false
	}
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Allow(f RequestFilter) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if f(r) {
				h(w, r)
			} else {
				// TODO
				w.WriteHeader(http.StatusForbidden)
			}
		}
	}
}

func SetHeader(key, value string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			f(w, r)
		}
	}
}

func Logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%v] - %s %s\n", time.Now(), r.Method, r.RequestURI)
		f(w, r)
	}
}

// now it's possible to call this:
// filteredHandler := Allow(CIDR("127.0.0.1/32"))(MyHandler)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello\n")
}

func main() {
	routes := Routes{
		"/hello": {
			Handler: hello,
			Middleware: Stack{
				Logging,
			},
		},
		"/private": {
			Handler: hello,
			Allow: Filters{
				CIDR("127.0.0.1/32"),
				PasswordHeader("opensesame"),
			},
			Middleware: Stack{
				Logging,
			},
		},
		"/test": {
			Handler: hello,
			Middleware: Stack{
				Logging,
				SetHeader("X-Foo", "Bar"),
			},
		},
	}
	log.Fatal(routes.Serve("localhost:1217"))
}

type Filters []RequestFilter

// Combine creates a RequestFilter that is the conjunction
// of all the RequestFilters in f.
func (f Filters) Combine() RequestFilter {
	return func(r *http.Request) bool {
		for _, filter := range f {
			if !filter(r) {
				return false
			}
		}
		return true
	}
}

type Stack []Middleware

// Apply returns an http.Handlerfunc that has had all of the
// Middleware functions in s, if any, to f.
func (s Stack) Apply(f http.HandlerFunc) http.HandlerFunc {
	g := f
	for _, middleware := range s {
		g = middleware(g)
	}
	return g
}

type Endpoint struct {
	Handler    http.HandlerFunc
	Allow      Filters
	Middleware Stack
}

// Builds the endpoint described by e, by applying
// access restrictions and other middleware.
func (e Endpoint) Build() http.HandlerFunc {
	allowFilter := e.Allow.Combine()
	restricted := Allow(allowFilter)(e.Handler)

	return e.Middleware.Apply(restricted)
}

var myEndpoint = Endpoint{
	Handler: hello,
	Allow: Filters{
		CIDR("127.0.0.1/32"),
	},
	Middleware: Stack{
		Logging,
		SetHeader("X-Foo", "Bar"),
	},
}

type Routes map[string]Endpoint

func (r Routes) Serve(addr string) error {
	mux := http.NewServeMux()
	for pattern, endpoint := range r {
		mux.Handle(pattern, endpoint.Build())
	}

	return http.ListenAndServe(addr, mux)
}
