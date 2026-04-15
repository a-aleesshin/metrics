package logger

import "fmt"

type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

type Field struct {
	Key   string
	Value any
}

func String(key, v string) Field {
	return Field{Key: key, Value: v}
}

func Int(key string, v int) Field {
	return Field{Key: key, Value: v}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Bool(key string, v bool) Field {
	return Field{Key: key, Value: v}
}

func Any(key string, v any) Field {
	return Field{Key: key, Value: v}
}

func DurationMillis(key string, ms int64) Field {
	return Field{Key: key, Value: fmt.Sprintf("%dms", ms)}
}
