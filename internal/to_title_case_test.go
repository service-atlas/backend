package internal

import "testing"

func Test_toTitleCase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "lowercase",
			in:   "database",
			want: "Database",
		},
		{
			name: "uppercase",
			in:   "API",
			want: "Api",
		},
		{
			name: "mixed case",
			in:   "wOrKeR",
			want: "Worker",
		},
		{
			name: "single character lowercase",
			in:   "a",
			want: "A",
		},
		{
			name: "single character uppercase",
			in:   "A",
			want: "A",
		},
		{
			name: "with spaces (only first letter capitalized)",
			in:   "web service",
			want: "Web Service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToTitleCase(tt.in); got != tt.want {
				t.Errorf("toTitleCase() = %q, want %q", got, tt.want)
			}
		})
	}
}
