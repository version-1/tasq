package tq

import (
	"encoding/json"
	"fmt"
	"io"
)

type cliError struct {
	message string
	code    int
}

func writeCLIError(w io.Writer, message string, code int) int {
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
	return code
}

func usageError(format string, args ...any) error {
	return cliError{message: fmt.Sprintf(format, args...), code: 2}
}

func (e cliError) Error() string {
	return e.message
}
