package tq

import (
	"fmt"
)

type cliError struct {
	message string
	code    int
}

func usageError(format string, args ...any) error {
	return cliError{message: fmt.Sprintf(format, args...), code: 2}
}

func (e cliError) Error() string {
	return e.message
}
