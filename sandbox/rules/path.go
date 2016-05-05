package rules

import (
	"github.com/dchest/uniuri"
	"net/http"
)

type PathRule struct {
	id       string
	path     string
	disabled bool
}

func NewPath(path string) *PathRule {
	return &PathRule{id: uniuri.New(), path: path}
}

func (p *PathRule) ID() string {
	return p.id
}

func (p *PathRule) Name() string {
	return "path"
}

func (p *PathRule) Description() string {
	return "Matches HTTP request URL path againts a given path pattern"
}

func (p *PathRule) Disable() {
	p.disabled = true
}

func (p *PathRule) IsEnabled() bool {
	return p.disabled
}

func (p *PathRule) JSONConfig() string {
	// testing!
	return "{\"path\": \"" + p.path + "\"}"
}

func (p *PathRule) Match(req *http.Request) bool {
	return req.URL.Path == p.path
}
