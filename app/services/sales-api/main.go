package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"github.com/farmani/service/app/services/sales-api/handlers"
	"github.com/farmani/service/business/web/v1/debug"
	"github.com/farmani/service/foundation/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

/*
Need to figure out timeouts for http service.
Add Category field and type to product.
*/
var build = "develop"

func main() {
	log, err := logger.NewZapLogger("sales-api")
	if err != nil {
		fmt.Println("Failed to initialize logger", err)
		os.Exit(1)
	}
	defer func(log *zap.SugaredLogger) {
		err := log.Sync()
		if err != nil {
			return
		}
	}(log)

	if err := run(log); err != nil {
		log.Errorw("Failed to run service", "Error", err)
		err := log.Sync()
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	// -------------------------------------------------
	// Start the service

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "RAMIN FARMANI",
		},
	}
	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.StandardLibraryMux()); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------
	// Start Application Service
	defer log.Infow("Shutdown complete")
	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Infow(
		"startup",
		"version", cfg.Version.Build,
		"env", "development",
		"port", cfg.Web.APIHost,
		"GOMAXPROCS", runtime.GOMAXPROCS(0),
	)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Build:    build,
		Log:      log,
		Shutdown: shutdown,
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			err := api.Close()
			if err != nil {
				return fmt.Errorf("could not close server gracefully: %w", err)
			}
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
