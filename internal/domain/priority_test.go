package domain

import "testing"

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input   string
		want    Priority
		wantErr bool
	}{
		{input: "low", want: PriorityLow},
		{input: "MEDIUM", want: PriorityMedium},
		{input: "  High  ", want: PriorityHigh},
		{input: "", want: PriorityMedium},
		{input: "urgent", wantErr: true},
		{input: "1", wantErr: true},
	}

	for _, test := range tests {
		got, err := ParsePriority(test.input)
		if test.wantErr {
			if err == nil {
				t.Fatalf("ParsePriority(%q) = %q, want error", test.input, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParsePriority(%q) returned error: %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("ParsePriority(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestPriorityRank(t *testing.T) {
	if !(PriorityHigh.Rank() < PriorityMedium.Rank() &&
		PriorityMedium.Rank() < PriorityLow.Rank()) {
		t.Fatal("priority rank should order high before medium before low")
	}
}
