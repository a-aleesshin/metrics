// Package logger реализует порт логгера поверх go.uber.org/zap.
package logger

import (
	"github.com/a-aleesshin/metrics/internal/shared/port/logger"
	"go.uber.org/zap"
)

// ZapLogger — адаптер порта logger.Logger поверх *zap.Logger.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger оборачивает базовый zap-логгер; при nil base паникует.
func NewZapLogger(base *zap.Logger) *ZapLogger {
	if base == nil {
		panic("logger: base zap logger is nil")
	}

	return &ZapLogger{
		logger: base,
	}
}

// Info пишет сообщение уровня Info с переданными полями.
func (z *ZapLogger) Info(msg string, fields ...logger.Field) {
	z.logger.Info(msg, zapFields(fields)...)
}

// Error пишет сообщение уровня Error с переданными полями.
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
