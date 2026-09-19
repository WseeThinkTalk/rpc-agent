package util

import (
	"testing"
	"time"
)

func TestGenerateCode(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 10; i++ {
		code := GenerateCode(6)
		if len(code) != 6 {
			t.Errorf("GenerateCode(6) length = %d, want 6", len(code))
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("GenerateCode contains non-digit: %c", c)
			}
		}
		codes[code] = true
	}

	// codes should have some variety (very unlikely all same)
	if len(codes) < 3 {
		t.Log("warning: generated codes had low variety, might be coincidence")
	}
}

func TestGenerateCodeDifferentSizes(t *testing.T) {
	tests := []int{4, 6, 8}
	for _, size := range tests {
		code := GenerateCode(size)
		if len(code) != size {
			t.Errorf("GenerateCode(%d) length = %d, want %d", size, len(code), size)
		}
	}
}

func TestEndOfDay(t *testing.T) {
	tm := time.Date(2026, 5, 26, 10, 30, 45, 123, time.UTC)
	end := EndOfDay(tm)

	if end.Year() != 2026 || end.Month() != 5 || end.Day() != 26 {
		t.Errorf("EndOfDay date mismatch: %v", end)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("EndOfDay time mismatch: %v", end)
	}
	if end.Nanosecond() != 0 {
		t.Errorf("EndOfDay nanosecond should be 0, got %d", end.Nanosecond())
	}
}

func TestEndOfDayLocation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	tm := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	end := EndOfDay(tm)

	if end.Location().String() != loc.String() {
		t.Errorf("EndOfDay should preserve location, got %s", end.Location())
	}
}
