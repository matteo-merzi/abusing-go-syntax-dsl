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
	fromLocalHost := Logging(Allow(CIDR("127.0.0.1/32"))(hello))
	http.HandleFunc("/hello", fromLocalHost)
	log.Fatal(http.ListenAndServe("0.0.0.0:1217", nil))
}
