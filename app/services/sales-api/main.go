package main

import (
	"fmt"
	"github.com/farmani/service/foundation/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var build = "develop"

func main() {
	log, err := logger.NewZapLogger("sales-api", "logs/sales-api.log")
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
