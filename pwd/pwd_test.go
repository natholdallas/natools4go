package pwd

import (
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
)

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
		// Parameters outside the accepted ranges (too little / too much work).
		"$argon2id$v=19$m=1,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=65536,t=99,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		// Bad salt/key lengths.
		"$argon2id$v=19$m=65536,t=3,p=2$QUJD$QUJD",
	}
	for _, h := range cases {
		if Verify("whatever", h) {
			t.Errorf("Verify(%q) = true, want false", h)
		}
	}
}

// TestVerifyAcceptsOtherTunedParameters is the regression test for the old
// behaviour of hard-coding m=65536,t=3,p=2: any future tuning of Hash would have
// silently invalidated every stored password. Verify now accepts a bounded range
// so hashes produced with different parameters still work.
func TestVerifyAcceptsOtherTunedParameters(t *testing.T) {
	const plain = "correct horse battery staple"

	salt := make([]byte, saltLen)
	for i := range salt {
		salt[i] = byte(i + 1)
	}

	// Same algorithm, different-but-acceptable cost parameters.
	const memory, timeCost, parallelism = 32 * 1024, 2, 1
	key := argon2.IDKey([]byte(plain), salt, timeCost, memory, parallelism, keyLen)
	encoded := "$argon2id$v=19$" +
		fmt.Sprintf("m=%d,t=%d,p=%d", memory, timeCost, parallelism) + "$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(key)

	if !Verify(plain, encoded) {
		t.Fatal("Verify rejected a valid hash produced with other parameters")
	}
	if Verify("wrong", encoded) {
		t.Fatal("Verify accepted a wrong password for the tuned hash")
	}
}
