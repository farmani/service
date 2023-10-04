package main

import (
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"github.com/farmani/service/foundation/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

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
	// Setup our application
	log.Infow(
		"Starting service",
		"version", "1.0.0",
		"env", "development",
		"port", "8080",
		"GOMAXPROCS", runtime.GOMAXPROCS(0),
	)

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

	// -------------------------------------------------
	// Start the service
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	log.Infow(
		"Shutdown signal received",
		"signal", sig.String(),
	)
	defer log.Infow("Shutdown complete")

	return nil
}
