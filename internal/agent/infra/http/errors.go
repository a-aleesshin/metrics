package httpadapter

import (
	"errors"
	"fmt"
)

type unexpectedStatusError struct {
	code int
}

func (e unexpectedStatusError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.code)
}

func isRetriableHTTPError(err error) bool {
	var statusErr unexpectedStatusError
	if errors.As(err, &statusErr) {
		return false
	}

	return err != nil
}
