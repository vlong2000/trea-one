package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vlong2000/trea-one/nfs/nrf/internal/sbi"
	"github.com/vlong2000/trea-one/nfs/nrf/internal/store"
	"github.com/vlong2000/trea-one/nfs/nrf/pkg/factory"
)

type App struct {
	cfg    *factory.Config
	server *sbi.Server
	store  *store.Store
	logger *log.Logger
}

func New(cfg *factory.Config) *App {
	logger := log.New(os.Stderr, "[nrf] ", log.LstdFlags)
	st := store.New()
	return &App{
		cfg:    cfg,
		store:  st,
		server: sbi.New(cfg, st, logger),
		logger: logger,
	}
}

func (a *App) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()
	a.logger.Printf("NRF started scheme=%s addr=%s", a.cfg.Scheme(), a.cfg.BindingAddr())
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.server.Shutdown(shutdownCtx)
		a.logger.Printf("NRF terminated")
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("nrf: serve: %w", err)
		}
		return nil
	}
}

func Run(cfgPath string) error {
	cfg, err := factory.ReadConfig(cfgPath)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return New(cfg).Start(ctx)
}
