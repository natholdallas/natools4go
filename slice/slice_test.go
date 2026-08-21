package slice

import "testing"

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(v int) int { return v * 2 })
	want := []int{2, 4, 6}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	if len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Errorf("got %v, want [2 4]", got)
	}
}

func TestForEach(t *testing.T) {
	sum := 0
	ForEach([]int{1, 2, 3}, func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("sum = %d, want 6", sum)
	}
}

func TestDefu(t *testing.T) {
	if got := Defu(5, nil); got != 5 {
		t.Errorf("Defu default = %d, want 5", got)
	}
	if got := Defu(5, []int{9}); got != 9 {
		t.Errorf("Defu arg = %d, want 9", got)
	}
}
