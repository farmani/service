package main

import (
	"errors"
	"fmt"
	"github.com/farmani/service/foundation/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	log := logger.NewZapLogger("logs/error.logs", "developement")

	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			fmt.Println("Failed to sync logger", err)
		}
	}(log)

	if err := run(log); err != nil {
		log.Error("Failed to run service", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.Logger) error {

	// -------------------------------------------------
	// Setup our application
	log.Info(
		"Starting service",
		zap.String("version", "1.0.0"),
		zap.String("env", "developement"),
		zap.String("port", "8080"),
		zap.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)),
	)

	// -------------------------------------------------
	// Start the service
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	log.Info("Shutdown signal received", zap.String("signal", sig.String()))
	defer log.Info("Shutdown complete")
	return nil
}
