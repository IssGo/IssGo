// Package logger provides structured logging via zap.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	Log = zap.NewNop().Sugar()
}

type Config struct {
	Level   string // debug, info, warn, error
	Verbose bool
	JSON    bool
}

func DefaultConfig() Config {
	return Config{Level: "info", Verbose: false, JSON: false}
}

func Init(cfg Config) {
	var zcfg zap.Config

	if cfg.Verbose || cfg.Level == "debug" {
		zcfg = zap.NewDevelopmentConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zcfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		zcfg = zap.NewProductionConfig()
		zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	if cfg.JSON {
		zcfg.Encoding = "json"
	}

	switch cfg.Level {
	case "debug":
		zcfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		zcfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zcfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zcfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	zcfg.DisableCaller = true
	zcfg.DisableStacktrace = !cfg.Verbose

	logger, err := zcfg.Build()
	if err != nil {
		panic("logger init: " + err.Error())
	}
	Log = logger.Sugar()
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
