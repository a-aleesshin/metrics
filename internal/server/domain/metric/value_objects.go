package metric

// ID — идентификатор метрики; непустая строка.
type ID string

// String возвращает идентификатор строкой.
func (i ID) String() string {
	return string(i)
}

// NewID создаёт ID; пустая строка — ErrIDEmpty.
func NewID(v string) (ID, error) {
	if v == "" {
		return "", ErrIDEmpty
	}

	return ID(v), nil
}

// Name — имя метрики; непустая строка.
type Name string

// String возвращает имя строкой.
func (n Name) String() string {
	return string(n)
}

// NewName создаёт Name; пустая строка — ErrNameEmpty.
func NewName(v string) (Name, error) {
	if v == "" {
		return "", ErrNameEmpty
	}

	return Name(v), nil
}
