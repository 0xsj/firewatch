// cmd/api/endpoints.go
package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/0xsj/hexagonal-go/pkg/observability/logger/console"
	"github.com/go-chi/chi/v5"
)

// Endpoint represents a discovered route.
type Endpoint struct {
	Method string
	Path   string
}

// EndpointGroup represents a group of endpoints by domain/prefix.
type EndpointGroup struct {
	Name      string
	Endpoints []Endpoint
}

// MethodColors maps HTTP methods to colors.
var MethodColors = map[string]string{
	"GET":     console.BrightGreen,
	"POST":    console.BrightYellow,
	"PUT":     console.BrightBlue,
	"PATCH":   console.BrightCyan,
	"DELETE":  console.BrightRed,
	"HEAD":    console.BrightMagenta,
	"OPTIONS": console.BrightBlack,
}

// GroupOrder defines the display order for endpoint groups.
var GroupOrder = []string{
	"system",
	"documentation",
	"identity",
	"tenants",
	"email",
	"flags",
	"permissions",
	"audit",
	"admin",
	"demo",
	"other",
}

// DiscoverEndpoints walks the chi router and extracts all routes.
func DiscoverEndpoints(r chi.Router) []Endpoint {
	var endpoints []Endpoint

	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Skip wildcard catch-all routes
		if strings.HasSuffix(route, "/*") {
			return nil
		}
		// Skip empty routes
		if route == "" {
			return nil
		}

		endpoints = append(endpoints, Endpoint{
			Method: method,
			Path:   route,
		})
		return nil
	})

	return endpoints
}

// GroupEndpoints groups endpoints by path prefix/domain.
func GroupEndpoints(endpoints []Endpoint) []EndpointGroup {
	groups := make(map[string][]Endpoint)

	for _, ep := range endpoints {
		groupName := inferGroup(ep.Path)
		groups[groupName] = append(groups[groupName], ep)
	}

	// Sort endpoints within each group
	for name := range groups {
		sort.Slice(groups[name], func(i, j int) bool {
			if groups[name][i].Path == groups[name][j].Path {
				return methodPriority(groups[name][i].Method) < methodPriority(groups[name][j].Method)
			}
			return groups[name][i].Path < groups[name][j].Path
		})
	}

	// Build ordered result
	var result []EndpointGroup
	for _, name := range GroupOrder {
		if eps, ok := groups[name]; ok && len(eps) > 0 {
			result = append(result, EndpointGroup{
				Name:      name,
				Endpoints: eps,
			})
			delete(groups, name)
		}
	}

	// Add any remaining groups not in GroupOrder
	var remaining []string
	for name := range groups {
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)
	for _, name := range remaining {
		if len(groups[name]) > 0 {
			result = append(result, EndpointGroup{
				Name:      name,
				Endpoints: groups[name],
			})
		}
	}

	return result
}

// inferGroup determines the group name from a path.
func inferGroup(path string) string {
	// Health check
	if path == "/health" {
		return "system"
	}

	// Swagger/docs
	if strings.HasPrefix(path, "/swagger") {
		return "documentation"
	}

	// Admin dashboard
	if strings.HasPrefix(path, "/admin/") {
		return "admin"
	}

	// Demo
	if strings.HasPrefix(path, "/demo") {
		return "demo"
	}

	// API routes
	if strings.HasPrefix(path, "/api/v1/") {
		segment := extractAPISegment(path)
		switch segment {
		case "users", "auth", "sessions":
			return "identity"
		case "tenants":
			return "tenants"
		case "email":
			return "email"
		case "flags":
			return "flags"
		case "permissions", "roles", "assignments":
			return "permissions"
		case "audit":
			return "audit"
		default:
			return segment
		}
	}

	return "other"
}

// extractAPISegment extracts the first path segment after /api/v1/.
func extractAPISegment(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(trimmed, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "other"
}

// methodPriority returns a sort priority for HTTP methods.
func methodPriority(method string) int {
	priorities := map[string]int{
		"GET":     1,
		"POST":    2,
		"PUT":     3,
		"PATCH":   4,
		"DELETE":  5,
		"HEAD":    6,
		"OPTIONS": 7,
	}
	if p, ok := priorities[method]; ok {
		return p
	}
	return 99
}

// colorMethod returns a colorized method string.
func colorMethod(method string) string {
	color, ok := MethodColors[method]
	if !ok {
		color = console.White
	}
	// Pad method to 7 chars for alignment
	padded := fmt.Sprintf("%-7s", method)
	return color + padded + console.Reset
}

// colorPath returns a colorized path string.
func colorPath(path string) string {
	// Highlight path parameters
	result := path
	if strings.Contains(path, "{") {
		// Color parameters differently
		result = strings.ReplaceAll(result, "{", console.BrightMagenta+"{")
		result = strings.ReplaceAll(result, "}", "}"+console.Reset)
	}
	return result
}

// groupTitle returns a formatted group title.
func groupTitle(name string) string {
	titles := map[string]string{
		"system":        "System",
		"documentation": "Documentation",
		"identity":      "Identity (Users & Auth)",
		"tenants":       "Tenants",
		"email":         "Email Templates",
		"flags":         "Feature Flags",
		"permissions":   "Permissions & Roles",
		"audit":         "Audit Trail",
		"admin":         "Admin Dashboard",
		"demo":          "Demo",
		"other":         "Other",
	}
	if title, ok := titles[name]; ok {
		return title
	}
	return strings.Title(name)
}

// PrintEndpoints prints all discovered endpoints with colors.
func PrintEndpoints(r chi.Router, port, metricsPort int) {
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	metricsURL := fmt.Sprintf("http://localhost:%d", metricsPort)

	endpoints := DiscoverEndpoints(r)
	groups := GroupEndpoints(endpoints)

	fmt.Println()
	fmt.Println(console.Bold + console.BrightWhite + "════════════════════════════════════════════════════════════════" + console.Reset)
	fmt.Println(console.Bold + console.BrightWhite + "  Hexagonal Go API - Available Endpoints" + console.Reset)
	fmt.Println(console.Bold + console.BrightWhite + "════════════════════════════════════════════════════════════════" + console.Reset)
	fmt.Println()

	for _, group := range groups {
		// Print group header
		fmt.Printf(console.Bold+console.BrightCyan+"  ▸ %s"+console.Reset+"\n", groupTitle(group.Name))
		fmt.Println()

		// Print endpoints
		for _, ep := range group.Endpoints {
			fmt.Printf("    %s  %s%s\n", colorMethod(ep.Method), baseURL, colorPath(ep.Path))
		}
		fmt.Println()
	}

	// Print observability section
	fmt.Printf(console.Bold + console.BrightCyan + "  ▸ Observability" + console.Reset + "\n")
	fmt.Println()
	fmt.Printf("    %s  %s/metrics\n", colorMethod("GET"), metricsURL)
	fmt.Println()

	fmt.Println(console.Bold + console.BrightWhite + "════════════════════════════════════════════════════════════════" + console.Reset)
	fmt.Println()
}
