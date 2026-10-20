package http

import (
	"testing"

	"guest-accounting/internal/domain"
)

func TestApplyGuestPatch(t *testing.T) {
	tests := []struct {
		name     string
		existing domain.Guest
		patch    domain.Guest
		want     domain.Guest
	}{
		{
			name:     "all fields applied",
			existing: domain.Guest{FullName: "old", Phone: "old", RoomNumber: "old", CheckIn: "old", CheckOut: "old", Note: "old"},
			patch:    domain.Guest{FullName: "Иван", Phone: "+7", RoomNumber: "101", CheckIn: "2026-01-01", CheckOut: "2026-01-05", Note: "note"},
			want:     domain.Guest{FullName: "Иван", Phone: "+7", RoomNumber: "101", CheckIn: "2026-01-01", CheckOut: "2026-01-05", Note: "note"},
		},
		{
			name:     "no fields applied",
			existing: domain.Guest{FullName: "old", Phone: "old"},
			patch:    domain.Guest{},
			want:     domain.Guest{FullName: "old", Phone: "old"},
		},
		{
			name:     "partial fields applied",
			existing: domain.Guest{FullName: "old", Phone: "old", RoomNumber: "old"},
			patch:    domain.Guest{FullName: "Иван"},
			want:     domain.Guest{FullName: "Иван", Phone: "old", RoomNumber: "old"},
		},
		{
			name:     "empty string is not applied",
			existing: domain.Guest{Phone: "old"},
			patch:    domain.Guest{Phone: ""},
			want:     domain.Guest{Phone: "old"},
		},
		{
			name:     "whitespace string is applied",
			existing: domain.Guest{Phone: "old"},
			patch:    domain.Guest{Phone: "  "},
			want:     domain.Guest{Phone: "  "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyGuestPatch(&tt.existing, tt.patch)
			if tt.existing != tt.want {
				t.Errorf("applyGuestPatch() = %+v, want %+v", tt.existing, tt.want)
			}
		})
	}
}

func TestSetIfNotEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"non-empty value", "hello", "hello"},
		{"empty value not applied", "", "old"},
		{"whitespace value applied", "  ", "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "old"
			setIfNotEmpty(&target, tt.value)
			if target != tt.want {
				t.Errorf("setIfNotEmpty(&%q, %q) = %q, want %q", target, tt.value, target, tt.want)
			}
		})
	}
}
