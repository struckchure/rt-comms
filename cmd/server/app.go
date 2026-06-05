package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/samber/do/v2"
	"github.com/struckchure/rt-comms/internal"
)

type server struct {
	e    *echo.Echo
	deps appDeps
	cfg  *internal.Config
}

func newServer(cfg *internal.Config) (*server, error) {
	injector := do.New()
	if err := register(injector, cfg); err != nil {
		return nil, err
	}
	deps := mustDeps(injector)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	e := echo.NewWithConfig(echo.Config{Logger: logger})
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", deps.Health.Health)
	e.GET("/ws", deps.WebSocket.HandleWS)
	e.GET("/users/:user/direct/:peer/messages", deps.History.ListMessages)
	e.Static("/static", "static")
	e.GET("/testclient.html", func(c *echo.Context) error {
		return c.File("static/testclient.html")
	})

	return &server{e: e, deps: deps, cfg: cfg}, nil
}

func (s *server) run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", s.cfg.Port)
		sc := echo.StartConfig{
			Address:         s.cfg.Port,
			GracefulTimeout: 10 * time.Second,
		}
		err := sc.Start(ctx, s.e)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.deps.Call.EndAll()
	_ = s.deps.Conn.Shutdown(shutdownCtx)
	if rs, ok := s.deps.Injector.(*do.RootScope); ok {
		rs.ShutdownWithContext(shutdownCtx)
	}
	s.deps.Pool.Close()

	if err := <-errCh; err != nil {
		return err
	}
	return nil
}
