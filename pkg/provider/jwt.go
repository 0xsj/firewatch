package provider

import (
	"github.com/0xsj/hexagonal-go/pkg/security/jwt"
)

// ProvideJWTService creates a JWT service.
func ProvideJWTService(config jwt.Config) jwt.Service {
	return jwt.NewService(config)
}
