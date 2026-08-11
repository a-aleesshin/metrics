// Package logger определяет порт структурированного логирования, независимый от реализации.
package logger

import "fmt"

// Logger — минимальный интерфейс структурированного логгера.
type Logger interface {
	// Info пишет сообщение уровня Info с переданными полями.
	Info(msg string, fields ...Field)
	// Error пишет сообщение уровня Error с переданными полями.
	Error(msg string, fields ...Field)
}

// Field — пара ключ-значение для структурированного лога.
type Field struct {
	Key   string
	Value any
}

// String создаёт поле со строковым значением.
func String(key, v string) Field {
	return Field{Key: key, Value: v}
}

// Int создаёт поле с целочисленным значением.
func Int(key string, v int) Field {
	return Field{Key: key, Value: v}
}

// Err создаёт поле "error" с текстом ошибки; для nil значение поля — nil.
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Bool создаёт поле с булевым значением.
func Bool(key string, v bool) Field {
	return Field{Key: key, Value: v}
}

// Any создаёт поле с произвольным значением.
func Any(key string, v any) Field {
	return Field{Key: key, Value: v}
}

// DurationMillis создаёт поле с длительностью в миллисекундах в виде строки "<N>ms".
func DurationMillis(key string, ms int64) Field {
	return Field{Key: key, Value: fmt.Sprintf("%dms", ms)}
}
