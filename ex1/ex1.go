package main

import (
	"fmt"
	"strings"
)

// TODO: abstraction on channels?

type Host string

type HostList []Host

func (l HostList) Select(f HostFilter) HostList {
	result := make(HostList, 0)
	for _, h := range l {
		if f(h) {
			result = append(result, h)
		}
	}
	return result
}

type HostFilter func(h Host) bool

func (f HostFilter) And(g HostFilter) HostFilter {
	return func(h Host) bool {
		return f(h) && g(h)
	}
}

func (f HostFilter) Or(g HostFilter) HostFilter {
	return func(h Host) bool {
		return f(h) || g(h)
	}
}

var IsDotOrg HostFilter = func(h Host) bool {
	return strings.HasSuffix(string(h), ".org")
}

var HasGo HostFilter = func(h Host) bool {
	return strings.Contains(string(h), "go")
}

var IsAcademy HostFilter = func(h Host) bool {
	return strings.Contains(string(h), "academy")
}

type HostSet map[Host]interface{}

func (s HostSet) Add(h Host) {
	s[h] = struct{}{}
}

func (s HostSet) Remove(h Host) {
	delete(s, h)
}

func (s HostSet) Contains(h Host) bool {
	_, found := s[h]
	return found
}

func main() {

	s := make(HostSet)
	s.Add("golang.org")
	s.Add("google.com")
	s.Add("gopheracademy.org")
	s.Remove("google.com")

	hostnames := HostList{
		"golang.org",
		"google.com",
		"gopheracademy.org",
	}

	for _, h := range hostnames {
		fmt.Printf("%s? %v\n", h, s.Contains(h))
	}
}
