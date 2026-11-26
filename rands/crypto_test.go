package rands

import (
	"strings"
	"testing"
)

func TestCryptoLengthAndAlphabet(t *testing.T) {
	for _, n := range []int{0, 1, 16, 32, 64, 1024} {
		got := Crypto(n)
		if len(got) != n {
			t.Errorf("Crypto(%d) length = %d, want %d", n, len(got), n)
		}
		for _, r := range got {
			if !strings.ContainsRune(alphabet, r) {
				t.Errorf("Crypto(%d) contains non-alphanumeric rune %q", n, r)
			}
		}
	}
	if Crypto(-1) != "" || Crypto(0) != "" {
		t.Error("Crypto should return an empty string for non-positive lengths")
	}
}

func TestCryptoIsNotRepeating(t *testing.T) {
	// A CSPRNG must not return the same 32-char value twice in a row. This is a
	// smoke test for "did someone accidentally swap in math/rand again".
	first := Crypto(32)
	second := Crypto(32)
	if first == second {
		t.Error("two Crypto(32) calls produced identical output")
	}
}

// TestCharIsCryptoSafe locks in the v1.27.0 behaviour change: Char used to be
// math/rand-backed, which silently made every secret generated with it
// predictable. It must now be an alias of Crypto.
func TestCharIsCryptoSafe(t *testing.T) {
	if len(Char(32)) != 32 {
		t.Error("Char(32) should still return 32 characters")
	}
	if Char(32) == CharFast(32) {
		// Different sources can collide by chance, but matching output for the
		// same length on back-to-back calls is astronomically unlikely.
		t.Log("Char matched CharFast, retrying once")
		if Char(32) == CharFast(32) {
			t.Error("Char appears to share CharFast's source; expected crypto/rand")
		}
	}
}

func TestHexAndCryptoDigits(t *testing.T) {
	h := Hex(32)
	if len(h) != 32 {
		t.Errorf("Hex(32) length = %d, want 32", len(h))
	}
	for _, r := range h {
		if !strings.ContainsRune(hexAlphabet, r) {
			t.Errorf("Hex produced non-hex rune %q", r)
		}
	}

	d := CryptoDigits(6)
	if len(d) != 6 {
		t.Errorf("CryptoDigits(6) length = %d, want 6", len(d))
	}
	for _, r := range d {
		if r < '0' || r > '9' {
			t.Errorf("CryptoDigits produced non-digit rune %q", r)
		}
	}
}

func TestCryptoPick(t *testing.T) {
	list := []string{"a", "b", "c"}
	for range 20 {
		if got := CryptoPick(list); !strings.Contains("abc", got) {
			t.Fatalf("CryptoPick returned %q", got)
		}
	}
	if got := CryptoPick([]int{}); got != 0 {
		t.Errorf("CryptoPick(empty) = %d, want 0", got)
	}
}
