package maths

import "testing"

func TestDivCeil(t *testing.T) {
	cases := []struct {
		a, b int64
		want int64
	}{
		{5, 2, 3},
		{4, 2, 2},
		{0, 2, 0},
		{5, 0, 0},
		{7, 3, 3},
		{10, 3, 4},
		{1, 10, 1},
	}
	for _, c := range cases {
		if got := DivCeil(c.a, c.b); got != c.want {
			t.Errorf("DivCeil(%d,%d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestDigits(t *testing.T) {
	cases := []struct {
		n    int64
		want []int64
	}{
		{0, []int64{0}},
		{123, []int64{1, 2, 3}},
		{-456, []int64{4, 5, 6}},
		{9, []int64{9}},
	}
	for _, c := range cases {
		got := Digits(c.n)
		if len(got) != len(c.want) {
			t.Errorf("Digits(%d) = %v, want %v", c.n, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("Digits(%d) = %v, want %v", c.n, got, c.want)
				break
			}
		}
	}
}
