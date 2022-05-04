package logs

import (
	"github.com/gonzispina/gokit/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockgen -source=logs.go -destination=logs_mock.go -package=logs Logger

// Logger interface to use in every mock
type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
}

type logger struct {
	namespace string
	zap       *zap.Logger
}

// Info logging level
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.zap.Info(msg, addTrackingID(ctx, fields...)...)
}

// Warn logging level
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.zap.Warn(msg, addTrackingID(ctx, fields...)...)
}

// Debug logging level
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.zap.Debug(msg, addTrackingID(ctx, fields...)...)
}

// Error logging level
func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.zap.Error(msg, addTrackingID(ctx, fields...)...)
}

// Fatal logging level
func (l *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.zap.Fatal(msg, addTrackingID(ctx, fields...)...)
}

// Log returns an instance of a logger with a new namespace
func (l *logger) Log(namespace string) Logger {
	_ = l.zap.Sync()
	return &logger{
		namespace: namespace,
		zap:       l.zap,
	}
}

// InitCustom logger with a custom configuration.
func InitCustom(c zap.Config, namespace string) (Logger, error) {
	z, err := c.Build()
	if err != nil {
		return nil, err
	}
	return &logger{
		namespace: namespace,
		zap:       z,
	}, nil
}

// InitDefault Function initializes a logger using uber-go/zap package in the application.
func InitDefault() Logger {
	conf := zap.Config{
		Encoding: "json",
		Level:    zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "file",
			MessageKey:   "msg",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.FullCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    false,
	}

	z, _ := conf.Build()
	z = z.WithOptions(zap.AddCallerSkip(1))

	return &logger{
		namespace: "default",
		zap:       z,
	}
}

// InitTest logger used for testing.
func InitTest() Logger {
	conf := zap.Config{
		Encoding: "json",
		Level:    zap.NewAtomicLevelAt(zapcore.ErrorLevel),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "file",
			MessageKey:   "msg",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.FullCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    false,
	}
	z, _ := conf.Build()
	z = z.WithOptions(zap.AddCallerSkip(1))
	return &logger{
		namespace: "default",
		zap:       z,
	}
}
