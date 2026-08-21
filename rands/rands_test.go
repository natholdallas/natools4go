package rands

import (
	"math/big"
	"testing"
)

func TestDistributeSum(t *testing.T) {
	v := Distribute(100, 7)
	if len(v) != 7 {
		t.Fatalf("len = %d, want 7", len(v))
	}
	sum := 0
	for _, x := range v {
		sum += x
	}
	if sum != 100 {
		t.Errorf("sum = %d, want 100", sum)
	}
}

func TestDistributeStrict(t *testing.T) {
	v := DistributeStrict(10, 4)
	if len(v) != 4 {
		t.Fatalf("len = %d, want 4", len(v))
	}
	sum := 0
	for _, x := range v {
		if x < 1 {
			t.Errorf("part %d < 1", x)
		}
		sum += x
	}
	if sum != 10 {
		t.Errorf("sum = %d, want 10", sum)
	}
}

func TestDigits(t *testing.T) {
	n := big.NewInt(123456789)
	got, err := Digits(n, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got < 100 || got > 999 {
		t.Errorf("got %d, want a 3-digit number", got)
	}
	if _, err := Digits(n, 20); err == nil {
		t.Error("expected out-of-range error, got nil")
	}
	if _, err := Digits(n, 0); err == nil {
		t.Error("expected error for length 0, got nil")
	}
}

func TestChar(t *testing.T) {
	s := Char(16)
	if len(s) != 16 {
		t.Errorf("len = %d, want 16", len(s))
	}
}

func TestPick(t *testing.T) {
	if got := Pick([]int{}); got != 0 {
		t.Errorf("Pick empty = %d, want 0", got)
	}
	if got := Pick([]int{7}); got != 7 {
		t.Errorf("Pick single = %d, want 7", got)
	}
}
