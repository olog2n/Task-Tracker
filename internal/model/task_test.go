package model

import "testing"

func TestFromString(t *testing.T) {
	tests := []struct {
		name    string     // Case name
		input   string     // Test argument
		want    TaskStatus // Expected result
		wantErr bool       // Expected error
	}{
		{"valid backlog", "backlog", StatusBacklog, false},
		{"valid in_progress", "in_progress", StatusInProgress, false},
		{"valid done", "done", StatusDone, false},
		{"valid cancelled", "cancelled", StatusCancelled, false},
		{"invalid unknown", "unknown", StatusBacklog, true},
		{"invalid empty", "", StatusBacklog, true},
		{"case insensitive", "BACKLOG", StatusBacklog, false},
		{"with spaces", " done ", StatusDone, false},
		{"mixed case", "In_PrOgReSs", StatusInProgress, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("FromString() = %v, want = %v", got, tt.want)
			}
		})
	}
}
