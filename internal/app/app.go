package app

import (
	"Shortener/internal/migrations"
	"Shortener/internal/repository"
	"Shortener/internal/service"
	"Shortener/internal/transport"
	"Shortener/pkg/logger"
	"Shortener/pkg/postgres"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wb-go/wbf/config"
	"go.uber.org/zap"
)

type App struct {
	ShortenerServer *transport.ShortenerServer
	cfg             *config.Config
	ctx             context.Context
	wg              sync.WaitGroup
	cancel          context.CancelFunc
}

func NewApp(cfg *config.Config, parentCtx context.Context) *App {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	db, err := postgres.NewPostgres(cfg)
	if err != nil {
		panic(err)
	}

	applied, err := migrations.RunMigrations(db.Master, "./migrations")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Failed to run migrations", zap.Error(err))
		panic(err)
	}
	if applied {
		logger.GetLoggerFromCtx(ctx).Info("Database migrations completed successfully")
	}

	repo := repository.NewShortenerRepository(ctx, db)
	srv := service.NewShortenerService(ctx, repo)
	server := transport.NewShortenerServer(ctx, srv, cfg)
	return &App{
		ShortenerServer: server,
		cfg:             cfg,
		ctx:             ctx,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	a.wg.Add(1)
	go func() {
		logger.GetLoggerFromCtx(a.ctx).Info("Server started on address", zap.Any("address", a.cfg.GetString("Host")+":"+a.cfg.GetString("Port")))
		defer a.wg.Done()
		if err := a.ShortenerServer.Run(); err != nil {
			errCh <- err
			a.cancel()
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		logger.GetLoggerFromCtx(a.ctx).Error("error running app", zap.Error(err))
		return err
	case <-a.ctx.Done():
		logger.GetLoggerFromCtx(a.ctx).Info("context done")
	}

	return nil
}
