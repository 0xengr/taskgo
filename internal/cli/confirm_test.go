package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   bool
		prompt string
	}{
		{name: "yes", input: "y\n", want: true},
		{name: "YES case and space", input: "  YES  \n", want: true},
		{name: "no", input: "n\n", want: false},
		{name: "empty line", input: "\n", want: false},
		{name: "end of input", input: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := confirm(strings.NewReader(test.input), &out, "Delete? [y/N] ")
			if err != nil {
				t.Fatalf("confirm returned error: %v", err)
			}
			if got != test.want {
				t.Fatalf("confirm(%q) = %v, want %v", test.input, got, test.want)
			}
			if !strings.Contains(out.String(), "Delete?") {
				t.Fatalf("prompt was not written: %q", out.String())
			}
		})
	}
}
