package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger  *zap.Logger
	defaultConfig = Config{
		Strict:           false,
		Production:       false,
		EnableStacktrace: true,
	}
	// Maps name to logger
	namedLoggersCache = make(map[string]*zap.Logger)
	mu                sync.RWMutex
)

type Config struct {
	Out              []string
	Strict           bool
	Production       bool
	EnableStacktrace bool
}

func NewLogger(cfg Config) error {
	builder := zap.NewProductionConfig()
	builder.DisableStacktrace = !cfg.EnableStacktrace
	builder.Development = false
	builder.Encoding = "console"

	builder.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if cfg.Strict {
		builder.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	if cfg.Production {
		builder.OutputPaths = cfg.Out
		builder.Encoding = "json"
		builder.DisableCaller = true
	}

	builder.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := builder.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

func Get() *zap.Logger {
	if globalLogger == nil {
		if err := NewLogger(defaultConfig); err != nil {
			panic(err)
		}
	}
	return globalLogger
}

func Named(name string) *zap.Logger {
	if globalLogger == nil {
		if err := NewLogger(defaultConfig); err != nil {
			panic(err)
		}
	}
	mu.RLock()
	l, ok := namedLoggersCache[name]
	if ok {
		mu.RUnlock()
		return l
	}
	mu.RUnlock()
	mu.Lock()
	namedLoggersCache[name] = globalLogger.Named(name)
	defer mu.Unlock()
	return namedLoggersCache[name]
}
