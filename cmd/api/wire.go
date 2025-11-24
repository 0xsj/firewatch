//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/cmd/api/config"
	"github.com/0xsj/hexagonal-go/internal/identity"
	"github.com/0xsj/hexagonal-go/pkg/provider"
)

// InitializeApp wires up the entire application.
func InitializeApp(cfg *config.AppConfig) (*App, func(), error) {
	wire.Build(
		// Extract config components for providers
		wire.FieldsOf(new(*config.AppConfig), "Database", "Logger", "Server"),

		// Infrastructure (from pkg/provider)
		provider.ProvideLogger,
		provider.ProvideDatabase,
		provider.ProvideEventBus,

		// Identity domain (from internal/identity)
		identity.IdentitySet,

		// Wire the App struct
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil, nil
}
