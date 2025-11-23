//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/0xsj/hexagonal-go/internal/identity"
	"github.com/0xsj/hexagonal-go/pkg/provider"
)

// InitializeApp wires up the entire application.
func InitializeApp() (*App, func(), error) {
	wire.Build(
		// Infrastructure (from pkg/provider)
		provider.ProvideLogger,
		provider.ProvideDatabase,

		// Identity domain (from internal/identity)
		identity.IdentitySet,

		// Wire the App struct
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil, nil
}
