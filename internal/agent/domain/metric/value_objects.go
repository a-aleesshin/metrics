package metric

// Name — непустое имя метрики.
type Name string

// NewName создаёт Name; пустое значение даёт ErrNameEmpty.
func NewName(value string) (Name, error) {
	if value == "" {
		return "", ErrNameEmpty
	}

	return Name(value), nil
}

// String возвращает имя как строку.
func (n Name) String() string {
	return string(n)
}
