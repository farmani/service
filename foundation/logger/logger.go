package logger

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
	"syscall"
)

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

func NewZapLogger(path string, env string) *zap.Logger {
	logger := zap.Must(zap.NewProduction())
	switch env {
	case "production":
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.NameKey = "logger"
		encoderCfg.CallerKey = "caller"
		encoderCfg.MessageKey = "message"
		encoderCfg.StacktraceKey = "stacktrace"
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stdout",
				path,
			},
			ErrorOutputPaths: []string{
				"stderr",
			},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}
		logger = zap.Must(config.Build())
	case "staging":
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.NameKey = "logger"
		encoderCfg.CallerKey = "caller"
		encoderCfg.MessageKey = "message"
		encoderCfg.StacktraceKey = "stacktrace"
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stdout",
				path,
			},
			ErrorOutputPaths: []string{
				"stderr",
			},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}
		logger = zap.Must(config.Build())
	case "testing":
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.NameKey = "logger"
		encoderCfg.CallerKey = "caller"
		encoderCfg.MessageKey = "message"
		encoderCfg.StacktraceKey = "stacktrace"
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stdout",
				path,
			},
			ErrorOutputPaths: []string{
				"stderr",
			},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}
		logger = zap.Must(config.Build())
	case "development":
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.NameKey = "logger"
		encoderCfg.CallerKey = "caller"
		encoderCfg.MessageKey = "message"
		encoderCfg.StacktraceKey = "stacktrace"
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

		config := zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stdout",
				path,
			},
			ErrorOutputPaths: []string{
				"stderr",
				path,
			},
			InitialFields: map[string]interface{}{
				"pid": os.Getpid(),
			},
		}
		logger = zap.Must(config.Build())
	}

	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}(logger)
	return logger
}
