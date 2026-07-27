package tq

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteCLIErrorForFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "text", format: "text", want: ansiRed + "Error:" + ansiReset + " issue not found\n"},
		{name: "json", format: "json", want: "{\"error\":\"issue not found\"}\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			if code := writeCLIErrorForFormat(&buf, test.format, "issue not found", 1); code != 1 {
				t.Fatalf("code=%d, want 1", code)
			}
			if got := buf.String(); got != test.want {
				t.Fatalf("output=%q, want %q", got, test.want)
			}
			if test.format == "json" && strings.Contains(buf.String(), "\x1b[") {
				t.Fatalf("JSON error contains ANSI: %q", buf.String())
			}
		})
	}
}
