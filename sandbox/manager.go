package sandbox

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/dchest/uniuri"
	"gopkg.in/vinxi/layer.v0"
	"gopkg.in/vinxi/vinxi.v0"
)

type Layer struct {
	layer *layer.Layer
}

type Options struct {
	Optional bool
}

type Rule interface {
	ID() string
	Name() string
	Description() string
	// Options() Options
	JSONConfig() string
	Match(*http.Request) bool
}

type Scope struct {
	disabled bool
	rules    []Rule
	plugins  *PluginLayer

	ID          string
	Name        string
	Description string
}

func NewScope(rules ...Rule) *Scope {
	return &Scope{ID: uniuri.New(), Name: "default", plugins: NewPluginLayer(), rules: rules}
}

func (s *Scope) UsePlugin(plugin Plugin) {
	s.plugins.Use(plugin)
}

func (s *Scope) AddRule(rules ...Rule) {
	s.rules = append(s.rules, rules...)
}

func (s *Scope) Rules() []Rule {
	return s.rules
}

func (s *Scope) Disable() {
	s.disabled = true
}

func (s *Scope) Enable() {
	s.disabled = false
}

func (s *Scope) HandleHTTP(h http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.disabled {
			h.ServeHTTP(w, r)
			return
		}

		for _, rule := range s.rules {
			if !rule.Match(r) {
				// Continue
				h.ServeHTTP(w, r)
				return
			}
		}

		s.plugins.Run(w, r, h)
	}
}

type Manager struct {
	Server   *http.Server
	instance *vinxi.Vinxi
	scopes   []*Scope
}

func Manage(instance *vinxi.Vinxi) *Manager {
	m := &Manager{instance: instance}
	instance.Layer.UsePriority("request", layer.Tail, m)
	return m
}

func (m *Manager) NewScope(rules ...Rule) *Scope {
	scope := NewScope(rules...)
	m.scopes = append(m.scopes, scope)
	return scope
}

func (m *Manager) HandleHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	next := h

	for _, scope := range a.scopes {
		next = http.HandlerFunc(scope.HandleHTTP(next))
	}

	next.ServeHTTP(w, r)
}

type JSONRule struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Config      string `json:"config,omitempty"`
}

type JSONPlugin struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

type JSONScope struct {
	ID      string       `json:"id"`
	Name    string       `json:"name,omitempty"`
	Rules   []JSONRule   `json:"rules,omitempty"`
	Plugins []JSONPlugin `json:"plugins,omitempty"`
}

func (m *Manager) ServeAndListen(opts ServerOptions) (*http.Server, error) {
	a.Server = NewServer(opts)

	m := pat.New()
	a.Server.Handler = m

	// Define route handlers
	m.Get("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "vinxi HTTP API manager "+Version)
	}))

	m.Get("/scopes", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf := &bytes.Buffer{}
		scopes := createScopes(a.scopes)

		err := json.NewEncoder(buf).Encode(scopes)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(buf.Bytes())
	}))

	m.Get("/scopes/:scope", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get(":scope")

		// Find scope by ID
		for _, scope := range a.scopes {
			if scope.ID == id {
				data, err := encodeJSON(createScope(scope))
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
					return
				}
				w.Write(data)
				return
			}
		}

		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))

	return a.Server, Listen(a.Server, opts)
}

func encodeJSON(data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	return buf.Bytes(), err
}

func createScope(scope *Scope) JSONScope {
	return JSONScope{
		ID:      scope.ID,
		Name:    scope.Name,
		Rules:   createRules(scope),
		Plugins: createPlugins(scope),
	}
}

func createScopes(scopes []*Scope) []JSONScope {
	buf := make([]JSONScope, len(scopes))
	for i, scope := range scopes {
		buf[i] = createScope(scope)
	}
	return buf
}

func createRules(scope *Scope) []JSONRule {
	rules := make([]JSONRule, len(scope.rules))
	for i, rule := range scope.rules {
		rules[i] = JSONRule{ID: rule.ID(), Name: rule.Name(), Description: rule.Description(), Config: rule.JSONConfig()}
	}
	return rules
}

func createPlugins(scope *Scope) []JSONPlugin {
	plugins := make([]JSONPlugin, scope.plugins.Len())
	for i, plugin := range scope.plugins.pool {
		plugins[i] = JSONPlugin{ID: plugin.ID(), Name: plugin.Name(), Description: plugin.Description(), Enabled: plugin.IsEnabled()}
	}
	return plugins
}
