package logger

import (
	"os"
	"upcycle-hub/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger
var S *zap.SugaredLogger

func Init(cfg *config.LogConfig) error {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	var ws zapcore.WriteSyncer
	if cfg.File != "" {
		f, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(f))
	} else {
		ws = zapcore.AddSync(os.Stdout)
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), ws, level)
	L = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	S = L.Sugar()
	return nil
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

func Debug(msg string, fields ...zap.Field) {
	if L != nil {
		L.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if L != nil {
		L.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if L != nil {
		L.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if L != nil {
		L.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if L != nil {
		L.Fatal(msg, fields...)
	}
}

func Debugf(template string, args ...interface{}) {
	if S != nil {
		S.Debugf(template, args...)
	}
}

func Infof(template string, args ...interface{}) {
	if S != nil {
		S.Infof(template, args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if S != nil {
		S.Warnf(template, args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if S != nil {
		S.Errorf(template, args...)
	}
}
