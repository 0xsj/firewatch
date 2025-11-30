package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
)

// Renderer handles HTML template rendering.
type Renderer struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	mu        sync.RWMutex
	debug     bool
}

// Option configures the Renderer.
type Option func(*Renderer)

// WithDebug enables debug mode (reloads templates on each render).
func WithDebug(debug bool) Option {
	return func(r *Renderer) {
		r.debug = debug
	}
}

// WithFuncMap adds custom template functions.
func WithFuncMap(funcMap template.FuncMap) Option {
	return func(r *Renderer) {
		for k, v := range funcMap {
			r.funcMap[k] = v
		}
	}
}

// NewRenderer creates a new template renderer.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		funcMap:   defaultFuncMap(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// LoadFS loads templates from an embedded filesystem.
// layoutPattern: glob pattern for layout templates (e.g., "layouts/*.html")
// pagePattern: glob pattern for page templates (e.g., "pages/*.html")
func (r *Renderer) LoadFS(fsys fs.FS, layoutPattern, pagePattern string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find layout files
	layoutFiles, err := fs.Glob(fsys, layoutPattern)
	if err != nil {
		return fmt.Errorf("failed to glob layout files: %w", err)
	}

	// Find page files
	pageFiles, err := fs.Glob(fsys, pagePattern)
	if err != nil {
		return fmt.Errorf("failed to glob page files: %w", err)
	}

	// Parse each page with layouts
	for _, page := range pageFiles {
		name := filepath.Base(page)
		name = name[:len(name)-len(filepath.Ext(name))] // Remove extension

		// Combine layouts + page
		files := append(layoutFiles, page)

		tmpl, err := template.New(name).Funcs(r.funcMap).ParseFS(fsys, files...)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		r.templates[name] = tmpl
	}

	return nil
}

// LoadDir loads templates from a directory.
func (r *Renderer) LoadDir(dir, layoutPattern, pagePattern string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	layoutFiles, err := filepath.Glob(filepath.Join(dir, layoutPattern))
	if err != nil {
		return fmt.Errorf("failed to glob layout files: %w", err)
	}

	pageFiles, err := filepath.Glob(filepath.Join(dir, pagePattern))
	if err != nil {
		return fmt.Errorf("failed to glob page files: %w", err)
	}

	for _, page := range pageFiles {
		name := filepath.Base(page)
		name = name[:len(name)-len(filepath.Ext(name))]

		files := append(layoutFiles, page)

		tmpl, err := template.New(name).Funcs(r.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		r.templates[name] = tmpl
	}

	return nil
}

// Register registers a single template.
func (r *Renderer) Register(name string, tmpl *template.Template) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates[name] = tmpl
}

// Render renders a template to a writer.
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	r.mu.RLock()
	tmpl, ok := r.templates[name]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	return tmpl.ExecuteTemplate(w, "layout", data)
}

// RenderPartial renders a template without the layout wrapper.
func (r *Renderer) RenderPartial(w io.Writer, name, block string, data any) error {
	r.mu.RLock()
	tmpl, ok := r.templates[name]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	return tmpl.ExecuteTemplate(w, block, data)
}

// RenderString renders a template to a string.
func (r *Renderer) RenderString(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := r.Render(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// HTML renders a template and writes it as an HTTP response.
func (r *Renderer) HTML(w http.ResponseWriter, status int, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return r.Render(w, name, data)
}

// HTMLPartial renders a partial template (for HTMX responses).
func (r *Renderer) HTMLPartial(w http.ResponseWriter, status int, name, block string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	return r.RenderPartial(w, name, block, data)
}

// defaultFuncMap returns the default template functions.
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"eq": func(a, b any) bool {
			return a == b
		},
		"neq": func(a, b any) bool {
			return a != b
		},
		"contains": func(s, substr string) bool {
			return len(substr) > 0 && len(s) >= len(substr) && containsString(s, substr)
		},
		"default": func(defaultVal, val any) any {
			if val == nil || val == "" || val == 0 {
				return defaultVal
			}
			return val
		},
		"json": func(v any) template.JS {
			return template.JS(fmt.Sprintf("%v", v))
		},
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
