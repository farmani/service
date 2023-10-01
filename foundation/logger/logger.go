package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
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

func NewZapLogger(service string, paths ...string) (*zap.SugaredLogger, error) {
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
	}
	config.InitialFields = map[string]interface{}{
		"service": service,
		"pid":     os.Getpid(),
	}
	config.OutputPaths = []string{"stdout"}
	if paths != nil {
		config.OutputPaths = append(config.OutputPaths, paths...)
	}
	config.ErrorOutputPaths = []string{"stderr"}
	if paths != nil {
		config.ErrorOutputPaths = append(config.ErrorOutputPaths, paths...)
	}

	logger, err := config.Build(zap.WithCaller(true))
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
