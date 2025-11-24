package config

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ============================================================================
// File Source
// ============================================================================

// FileSource loads configuration from a file (YAML, JSON).
type FileSource struct {
	path    string
	options *FileSourceOptions
}

// NewFileSource creates a file-based configuration source with default options.
func NewFileSource(path string) *FileSource {
	return &FileSource{
		path:    path,
		options: DefaultFileSourceOptions(),
	}
}

// NewFileSourceWithOptions creates a file source with custom options.
func NewFileSourceWithOptions(path string, opts *FileSourceOptions) *FileSource {
	return &FileSource{
		path:    path,
		options: opts,
	}
}

func (s *FileSource) Name() string {
	return fmt.Sprintf("file:%s", s.path)
}

func (s *FileSource) Load(ctx context.Context, target any) error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file is required, return error; otherwise treat as missing source
			if s.options.Required {
				return fmt.Errorf("required config file not found: %s", s.path)
			}
			return &MissingSourceError{Source: s.Name()}
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine format
	format := s.options.Format
	if format == "" {
		// Detect from extension
		format = detectFormat(s.path)
	}

	// Unmarshal based on format
	switch format {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %w", err)
		}
	case "json":
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// detectFormat detects file format from extension.
func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return "yaml" // Default to YAML
	}
}

// ============================================================================
// Environment Source
// ============================================================================

// EnvSource loads configuration from environment variables.
type EnvSource struct {
	options *EnvSourceOptions
}

// NewEnvSource creates an environment variable source with default options.
func NewEnvSource(prefix string) *EnvSource {
	opts := DefaultEnvSourceOptions()
	opts.Prefix = prefix
	return &EnvSource{options: opts}
}

// NewEnvSourceWithOptions creates an environment source with custom options.
func NewEnvSourceWithOptions(opts *EnvSourceOptions) *EnvSource {
	return &EnvSource{options: opts}
}

func (s *EnvSource) Name() string {
	if s.options.Prefix != "" {
		return fmt.Sprintf("env:%s_*", s.options.Prefix)
	}
	return "env:*"
}

func (s *EnvSource) Load(ctx context.Context, target any) error {
	return loadFromEnv(target, s.options)
}

// loadFromEnv loads environment variables into target based on `env` tags.
func loadFromEnv(target any, opts *EnvSourceOptions) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	return setEnvValues(rv.Elem(), opts)
}

func setEnvValues(rv reflect.Value, opts *EnvSourceOptions) error {
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get env tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			// Recurse into nested structs
			if field.Kind() == reflect.Struct {
				if err := setEnvValues(field, opts); err != nil {
					return err
				}
			}
			continue
		}

		// Build full env var name with prefix
		envKey := envTag
		if opts.Prefix != "" {
			envKey = opts.Prefix + "_" + envTag
		}

		// Handle case sensitivity
		if !opts.CaseSensitive {
			envKey = strings.ToUpper(envKey)
		}

		// Get value from environment
		envValue, exists := os.LookupEnv(envKey)
		if !exists {
			continue
		}

		// Set the value
		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("field %s (env %s): %w", fieldType.Name, envKey, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	// Check if the field implements encoding.TextUnmarshaler
	if field.CanAddr() {
		if unmarshaler, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText([]byte(value))
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special case: time.Duration
		if field.Type().String() == "time.Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			field.SetInt(int64(duration))
			return nil
		}

		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}

	return nil
}

// ============================================================================
// Static Source (for testing/defaults)
// ============================================================================

// StaticSource provides static configuration (useful for testing).
type StaticSource struct {
	name   string
	config any
}

// NewStaticSource creates a static configuration source.
func NewStaticSource(name string, config any) *StaticSource {
	return &StaticSource{
		name:   name,
		config: config,
	}
}

func (s *StaticSource) Name() string {
	return fmt.Sprintf("static:%s", s.name)
}

func (s *StaticSource) Load(ctx context.Context, target any) error {
	// Deep copy config into target
	// This is a simplified version - in production, use a proper deep copy library
	data, err := yaml.Marshal(s.config)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, target)
}

// ============================================================================
// Errors
// ============================================================================

// MissingSourceError indicates a config source doesn't exist.
type MissingSourceError struct {
	Source string
}

func (e *MissingSourceError) Error() string {
	return fmt.Sprintf("config source not found: %s", e.Source)
}

func isMissingSourceError(err error) bool {
	_, ok := err.(*MissingSourceError)
	return ok
}
