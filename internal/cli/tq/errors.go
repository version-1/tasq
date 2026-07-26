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

func writeCLIErrorForFormat(w io.Writer, format string, message string, code int) int {
	if format == "json" {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		return code
	}
	_, _ = fmt.Fprintf(w, "%sError:%s %s\n", ansiRed, ansiReset, message)
	return code
}

func usageError(format string, args ...any) error {
	return cliError{message: fmt.Sprintf(format, args...), code: 2}
}

func (e cliError) Error() string {
	return e.message
}
