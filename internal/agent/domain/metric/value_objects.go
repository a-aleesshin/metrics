package metric

type Name string

func NewName(value string) (Name, error) {
	if value == "" {
		return "", ErrNameEmpty
	}

	return Name(value), nil
}

func (n Name) String() string {
	return string(n)
}
