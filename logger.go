package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger  *zap.SugaredLogger
	defaultConfig = Config{
		Strict:     false,
		Production: false,
	}
	// Maps name to logger
	namedLoggersCache = make(map[string]*zap.SugaredLogger)
	mu                sync.RWMutex
)

const (
	EncodingConsole = "console"
	EncodingJSON    = "json"
)

type Config struct {
	Out        []string
	Strict     bool
	Production bool
}

func NewLogger(cfg Config) error {
	builder := zap.NewProductionConfig()
	builder.DisableStacktrace = true
	builder.DisableCaller = false
	builder.Development = false
	builder.Encoding = EncodingConsole

	builder.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if cfg.Strict {
		builder.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	if cfg.Production {
		builder.OutputPaths = cfg.Out
		builder.Encoding = EncodingJSON
	}

	builder.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := builder.Build()
	if err != nil {
		return err
	}

	globalLogger = logger.Sugar()
	return nil
}

func Get() *zap.SugaredLogger {
	if globalLogger == nil {
		if err := NewLogger(defaultConfig); err != nil {
			panic(err)
		}
	}
	return globalLogger
}

func Named(name string) *zap.SugaredLogger {
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
