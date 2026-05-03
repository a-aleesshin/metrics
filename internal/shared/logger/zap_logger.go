package logger

import (
	"github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(base *zap.Logger) *ZapLogger {
	if base == nil {
		panic("logger: base zap logger is nil")
	}

	return &ZapLogger{
		logger: base,
	}
}

func (z *ZapLogger) Info(msg string, fields ...logger.Field) {
	z.logger.Info(msg, zapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...logger.Field) {
	z.logger.Error(msg, zapFields(fields)...)
}

func zapFields(fields []logger.Field) []zap.Field {
	out := make([]zap.Field, 0, len(fields))

	for _, field := range fields {
		out = append(out, zap.Any(field.Key, field.Value))
	}

	return out
}
