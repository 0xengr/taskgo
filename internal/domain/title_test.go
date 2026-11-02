package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	long := strings.Repeat("a", MaxTitleRunes)
	tooLong := strings.Repeat("a", MaxTitleRunes+1)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims surrounding space", input: "  fix it  ", want: "fix it"},
		{name: "collapses tabs and newlines", input: "fix\t \n it", want: "fix it"},
		{name: "collapses unicode whitespace", input: "fix\u00a0it", want: "fix it"},
		{name: "keeps inner punctuation", input: "fix: the (flaky) test", want: "fix: the (flaky) test"},
		{name: "at the rune limit", input: long, want: long},
		{name: "over the rune limit", input: tooLong, wantErr: true},
		{name: "control character", input: "fix\x00it", wantErr: true},
		{name: "empty", input: "   ", wantErr: true},
		{name: "invalid utf8", input: string([]byte{0xff, 0xfe}), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeTitle(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("NormalizeTitle(%q) = %q, want error", test.input, got)
				}
				var validation *ValidationError
				if !errors.As(err, &validation) {
					t.Fatalf("error %T is not a ValidationError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeTitle(%q) returned error: %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("NormalizeTitle(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestBuildTaskDescription(t *testing.T) {
	task, err := BuildTask("title", "", PriorityMedium)
	if err != nil {
		t.Fatalf("BuildTask returned error: %v", err)
	}
	if task.Description != nil {
		t.Fatalf("empty description should be absent, got %q", *task.Description)
	}

	task, err = BuildTask("title", "  keep me  ", PriorityMedium)
	if err != nil {
		t.Fatalf("BuildTask returned error: %v", err)
	}
	if task.Description == nil || *task.Description != "keep me" {
		t.Fatalf("description = %v, want %q", task.Description, "keep me")
	}
}
