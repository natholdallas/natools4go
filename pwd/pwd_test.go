package pwd

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := Hash("123456")
	if err != nil {
		t.Fatal(err)
	}
	if !IsHashed(h) {
		t.Fatalf("expected hashed prefix, got %q", h)
	}
	if !Verify("123456", h) {
		t.Fatal("expected correct password to verify")
	}
	if Verify("wrong", h) {
		t.Fatal("expected wrong password to fail")
	}
	if Verify("123456", "plaintext") {
		t.Fatal("expected plaintext to fail verify")
	}
	if Verify("123456", "") {
		t.Fatal("expected empty to fail verify")
	}
}

func TestVerifyMaliciousParamsNoPanic(t *testing.T) {
	// Attacker-crafted hashes with t=0 / p=0 used to crash argon2.IDKey
	// (which panics when time<1 or threads<1). They must now fail cleanly.
	cases := []string{
		"$argon2id$v=19$m=65536,t=0,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=65536,t=3,p=0$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		// Wrong parameters (non-standard m/t/p) must be rejected too.
		"$argon2id$v=19$m=1,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		// Bad salt/key lengths.
		"$argon2id$v=19$m=65536,t=3,p=2$QUJD$QUJD",
	}
	for _, h := range cases {
		if Verify("whatever", h) {
			t.Errorf("Verify(%q) = true, want false", h)
		}
	}
}
