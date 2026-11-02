package domain

import (
	"strings"
	"testing"
)

func TestNormalizeProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims", input: "  API  ", want: "API"},
		{name: "preserves case", input: "Home Lab", want: "Home Lab"},
		{name: "preserves inner spacing", input: "Home  Lab", want: "Home  Lab"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "control character", input: "bad\x00name", wantErr: true},
		{name: "too long", input: strings.Repeat("a", MaxProjectNameRunes+1), wantErr: true},
		{name: "invalid utf8", input: string([]byte{0xff}), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, key, err := NormalizeProjectName(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("NormalizeProjectName(%q) = %q/%q, want error", test.input, got, key)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeProjectName(%q) returned error: %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("name = %q, want %q", got, test.want)
			}
			if key == "" {
				t.Fatal("key is empty")
			}
		})
	}
}

func TestProjectNameKeysFold(t *testing.T) {
	pairs := [][2]string{
		{"API", "api"},
		{"STRASSE", "straße"},
		{"Café", "Cafe\u0301"},
	}

	for _, pair := range pairs {
		_, first, err := NormalizeProjectName(pair[0])
		if err != nil {
			t.Fatalf("NormalizeProjectName(%q): %v", pair[0], err)
		}
		_, second, err := NormalizeProjectName(pair[1])
		if err != nil {
			t.Fatalf("NormalizeProjectName(%q): %v", pair[1], err)
		}
		if first != second {
			t.Fatalf("keys differ: %q -> %q, %q -> %q", pair[0], first, pair[1], second)
		}
	}
}
