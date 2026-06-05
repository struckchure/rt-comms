package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/handlers"
	"github.com/struckchure/rt-comms/internal"
	"github.com/struckchure/rt-comms/repository"
	"github.com/struckchure/rt-comms/services"
)

type appDeps struct {
	Health    handlers.IHealthHandler
	History   handlers.IHistoryHandler
	WebSocket handlers.IWebSocketHandler
	Chat      services.IChatService
	Conn      services.IConnectionService
	Call      services.ICallService
	Pool      *pgxpool.Pool
	Injector  do.Injector
}

func register(injector do.Injector, cfg *internal.Config) error {
	do.ProvideValue(injector, cfg)
	do.Provide(injector, internal.NewPool)
	do.Provide(injector, repository.NewMessageRepository)
	do.Provide(injector, internal.NewHub)
	do.Provide(injector, services.NewCallSessionStore)
	do.Lazy(services.NewChatService)(injector)
	do.Lazy(services.NewCallService)(injector)
	do.Lazy(services.NewConnectionService)(injector)
	do.Provide(injector, handlers.NewHealthHandler)
	do.Provide(injector, handlers.NewHistoryHandler)
	do.Provide(injector, handlers.NewWebSocketHandler)
	return nil
}

func mustDeps(i do.Injector) appDeps {
	return appDeps{
		Health:    do.MustInvoke[handlers.IHealthHandler](i),
		History:   do.MustInvoke[handlers.IHistoryHandler](i),
		WebSocket: do.MustInvoke[handlers.IWebSocketHandler](i),
		Chat:      do.MustInvoke[services.IChatService](i),
		Conn:      do.MustInvoke[services.IConnectionService](i),
		Call:      do.MustInvoke[services.ICallService](i),
		Pool:      do.MustInvoke[*pgxpool.Pool](i),
		Injector:  i,
	}
}
