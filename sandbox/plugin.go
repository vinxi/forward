package sandbox

import (
	"net/http"

	"github.com/dchest/uniuri"
	"gopkg.in/vinxi/layer.v0"
)

type Handler func(http.Handler) http.Handler

// Plugin represents the required interface implemented by plugins.
type Plugin interface {
	// ID is used to retrieve the plugin unique identifier.
	ID() string
	// Name is used to retrieve the plugin name identifier.
	Name() string
	// Description is used to retrieve a human friendly
	// description of what the plugin does.
	Description() string
	// Enable is used to enable the current plugin.
	// If the plugin has been already enabled, the call is no-op.
	Enable()
	// Disable is used to disable the current plugin.
	Disable()
	// Remove is used to disable and remove a plugin.
	// Remove()
	// IsEnabled is used to check if a plugin is enabled or not.
	IsEnabled() bool
	// HandleHTTP is used to run the plugin task.
	// Note: add erro reporting layer
	HandleHTTP(http.Handler) http.Handler
}

type plugin struct {
	disabled    bool
	id          string
	name        string
	description string
	handler     Handler
}

func NewPlugin(name, description string, handler Handler) Plugin {
	return &plugin{id: uniuri.New(), name: name, description: description, handler: handler}
}

func (p *plugin) ID() string {
	return p.id
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Description() string {
	return p.description
}

func (p *plugin) Disable() {
	p.disabled = true
}

func (p *plugin) Enable() {
	p.disabled = false
}

func (p *plugin) IsEnabled() bool {
	return p.disabled == false
}

func (p *plugin) HandleHTTP(h http.Handler) http.Handler {
	next := p.handler(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// PluginLayer represents a plugins layer designed to intrument
// proxies providing plugin based dynamic configuration
// capabilities, such as register/unregister or
// enable/disable plugins at runtime satefy.
type PluginLayer struct {
	pool []Plugin
}

// NewPluginLayer creates a new plugins layer.
func NewPluginLayer() *PluginLayer {
	return &PluginLayer{}
}

func (l *PluginLayer) Use(plugin Plugin) {
	l.pool = append(l.pool, plugin)
}

func (l *PluginLayer) Len() int {
	return len(l.pool)
}

// Register implements the middleware Register method.
func (l *PluginLayer) Register(mw *layer.Layer) {
	mw.Use("error", l.Run)
	mw.Use("request", l.Run)
}

func (l *PluginLayer) Run(w http.ResponseWriter, r *http.Request, h http.Handler) {
	next := h
	for _, plugin := range l.pool {
		next = plugin.HandleHTTP(next)
	}
	next.ServeHTTP(w, r)
}
