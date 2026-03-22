package metric

type ID string

func (i ID) String() string {
	return string(i)
}

func NewID(v string) (ID, error) {
	if v == "" {
		return "", ErrIDEmpty
	}

	return ID(v), nil
}

type Name string

func (n Name) String() string {
	return string(n)
}

func NewName(v string) (Name, error) {
	if v == "" {
		return "", ErrNameEmpty
	}

	return Name(v), nil
}
