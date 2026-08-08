package metric

import "errors"

var (
	// ErrNameEmpty возвращается при попытке создать метрику с пустым именем.
	ErrNameEmpty = errors.New("name is empty")
)
