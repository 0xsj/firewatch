package domain

import "fmt"

// VariableType represents the type of a template variable.
type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeNumber  VariableType = "number"
	VariableTypeDate    VariableType = "date"
	VariableTypeURL     VariableType = "url"
	VariableTypeBoolean VariableType = "boolean"
)

// String returns the string representation of the variable type.
func (vt VariableType) String() string {
	return string(vt)
}

// Validate validates the variable type.
func (vt VariableType) Validate() error {
	switch vt {
	case VariableTypeString, VariableTypeNumber, VariableTypeDate, VariableTypeURL, VariableTypeBoolean:
		return nil
	default:
		return fmt.Errorf("invalid variable type: %s", vt)
	}
}

// IsValid returns true if the variable type is valid.
func (vt VariableType) IsValid() bool {
	return vt.Validate() == nil
}

// ParseVariableType parses a string into a VariableType.
func ParseVariableType(s string) (VariableType, error) {
	vt := VariableType(s)
	if err := vt.Validate(); err != nil {
		return "", err
	}
	return vt, nil
}

// Variable represents a placeholder variable in an email template.
type Variable struct {
	Name        string       `json:"name"`
	Type        VariableType `json:"type"`
	Required    bool         `json:"required"`
	Default     *string      `json:"default,omitempty"`
	Description string       `json:"description,omitempty"`
}

// Validate validates the variable.
func (v Variable) Validate() error {
	if v.Name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}
	if err := v.Type.Validate(); err != nil {
		return err
	}
	return nil
}

// NewVariable creates a new Variable.
func NewVariable(name string, varType VariableType, required bool, defaultVal *string, description string) (Variable, error) {
	v := Variable{
		Name:        name,
		Type:        varType,
		Required:    required,
		Default:     defaultVal,
		Description: description,
	}
	if err := v.Validate(); err != nil {
		return Variable{}, err
	}
	return v, nil
}

// Variables represents a collection of template variables.
type Variables []Variable

// Validate validates all variables in the collection.
func (vars Variables) Validate() error {
	seen := make(map[string]bool)
	for _, v := range vars {
		if err := v.Validate(); err != nil {
			return err
		}
		if seen[v.Name] {
			return fmt.Errorf("duplicate variable name: %s", v.Name)
		}
		seen[v.Name] = true
	}
	return nil
}

// Names returns the names of all variables.
func (vars Variables) Names() []string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	return names
}

// RequiredNames returns the names of all required variables.
func (vars Variables) RequiredNames() []string {
	var names []string
	for _, v := range vars {
		if v.Required {
			names = append(names, v.Name)
		}
	}
	return names
}

// Get returns a variable by name.
func (vars Variables) Get(name string) (Variable, bool) {
	for _, v := range vars {
		if v.Name == name {
			return v, true
		}
	}
	return Variable{}, false
}
