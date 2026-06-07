package csvparser

import (
	"strings"
	"testing"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantHeaders []string
		wantRows    int
		wantErr     bool
	}{
		{
			name:        "valid csv",
			input:       "name,age,city\nAlice,30,Lagos\nBob,25,Abuja",
			wantHeaders: []string{"name", "age", "city"},
			wantRows:    2,
		},
		{
			name:        "header only",
			input:       "name,age,city",
			wantHeaders: []string{"name", "age", "city"},
			wantRows:    0,
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: true,
		},
		{
			name:    "malformed csv — unclosed quote",
			input:   "name,age\n\"Alice,30",
			wantErr: true,
		},
		{
			name:        "single column",
			input:       "email\nuser@example.com",
			wantHeaders: []string{"email"},
			wantRows:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCSV(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Headers) != len(tt.wantHeaders) {
				t.Errorf("headers: got %v, want %v", result.Headers, tt.wantHeaders)
			}
			if len(result.Rows) != tt.wantRows {
				t.Errorf("row count: got %d, want %d", len(result.Rows), tt.wantRows)
			}
		})
	}
}
