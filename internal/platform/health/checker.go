package health

import "context"

// Checker — одна именованная проверка здоровья компонента.
type Checker interface {
	// Name возвращает имя проверки.
	Name() string
	// Check выполняет проверку; nil означает, что компонент здоров.
	Check(ctx context.Context) error
}
