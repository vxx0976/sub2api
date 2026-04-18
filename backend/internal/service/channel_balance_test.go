package service

import (
	"math"
	"testing"
)

func TestExtractBalanceValue(t *testing.T) {
	body := []byte(`{"data":{"balance":123.45,"limit":100,"used":30,"credits":9900}}`)

	cases := []struct {
		name    string
		path    string
		want    float64
		wantErr bool
	}{
		{"single path", "data.balance", 123.45, false},
		{"subtraction expression", "data.limit - data.used", 70, false},
		{"addition expression", "data.limit + data.used", 130, false},
		{"division by literal", "data.credits / 100", 99, false},
		{"multiplication", "data.used * 2", 60, false},
		{"missing path", "data.nonexistent", 0, true},
		{"missing operand in expr", "data.limit - data.missing", 0, true},
		{"division by zero literal", "data.balance / 0", 0, true},
		{"empty path - not a number", "", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractBalanceValue(body, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil (value=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractBalanceValueNumericRoot(t *testing.T) {
	got, err := extractBalanceValue([]byte(`42.5`), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42.5 {
		t.Fatalf("got %v, want 42.5", got)
	}
}
