package main

import (
	v1 "github.com/0xsj/hexagonal-go/internal/identity/interface/http/v1"
	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/messaging"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// App holds all application dependencies.
type App struct {
	Logger          logger.Logger
	DB              database.DB
	EventBus        messaging.Publisher // ← Add this!
	IdentityHandler *v1.Handler
}
