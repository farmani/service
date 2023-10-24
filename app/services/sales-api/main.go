package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/farmani/service/app/services/sales-api/handlers"
	database "github.com/farmani/service/business/sys/database/pgx"
	"github.com/farmani/service/business/web/auth"
	"github.com/farmani/service/business/web/v1/debug"
	"github.com/farmani/service/foundation/keystore"
	"github.com/farmani/service/foundation/logger"
	"go.uber.org/zap"
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
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service.sales-system.svc.cluster.local"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Auth struct {
			KeysFolder string `conf:"default:zarf/keys/"`
			ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			Issuer     string `conf:"default:zarf.sales.api"`
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
	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// -------------------------------------------------------------------------
	// Database support
	log.Infow("startup", "status", "initializing database support", "host", cfg.DB.Host)

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()
	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux(build, log, db)); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Infow("startup", "status", "initializing authentication.rego support")

	// Simple keystore versus using Vault.
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	authCfg := auth.Config{
		Log:       log,
		KeyLookup: ks,
	}

	authentication, err := auth.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing authentication: %w", err)
	}

	// -------------------------------------------------
	// Start Application Service
	defer log.Infow("Shutdown complete")

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
		Auth:     authentication,
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
