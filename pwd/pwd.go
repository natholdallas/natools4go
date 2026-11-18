// Package pwd provides argon2id password hashing helpers.
package pwd

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// hashPrefix identifies an argon2id hash at the current version. The cost
	// parameters that follow are per-hash, so IsHashed only matches this much.
	hashPrefix = "$argon2id$v=19$"
	saltLen    = 16
	keyLen     = 32
)

// Accepted parameter ranges for Verify. Hash always emits the constants below;
// the ranges let stored hashes from an older or future tuning keep verifying
// while still bounding the work an attacker can force through a crafted hash.
const (
	// Parameters Hash currently emits.
	hashMemory      = 64 * 1024
	hashTime        = 3
	hashParallelism = 2

	minMemory   = 8 * 1024
	maxMemory   = 1024 * 1024
	minTime     = 1
	maxTime     = 10
	minParallel = 1
	maxParallel = 16

	minSaltLen = 8
	maxSaltLen = 64
	minKeyLen  = 16
	maxKeyLen  = 128
)

// Hash derives an argon2id hash from the given plaintext password.
func Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(plain), salt, hashTime, hashMemory, hashParallelism, keyLen)
	return fmt.Sprintf("%sm=%d,t=%d,p=%d$", hashPrefix, hashMemory, hashTime, hashParallelism) +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(key), nil
}

func TryHash(plain string) string {
	hash, err := Hash(plain)
	if err != nil {
		return ""
	}
	return hash
}

// Verify reports whether the plaintext password matches the stored hash.
func Verify(plain, encoded string) bool {
	if !IsHashed(encoded) {
		return false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}
	var memory, timeCost, parallelism int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &parallelism); err != nil {
		return false
	}
	// Bound the parameters instead of requiring the exact current values, so a
	// future tuning of Hash does not invalidate every stored password. The lower
	// bounds stop argon2.IDKey panics on attacker-crafted hashes (t<1, p<1,
	// m<8*par) and the upper bounds block memory-exhaustion DoS.
	if memory < minMemory || memory > maxMemory {
		return false
	}
	if timeCost < minTime || timeCost > maxTime {
		return false
	}
	if parallelism < minParallel || parallelism > maxParallel {
		return false
	}
	// argon2 requires m >= 8*p; reject values that would overflow the internal
	// block computation.
	if memory < 8*parallelism {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	// Argon2 output length is not bounded by the algorithm, but anything wildly
	// larger than our own key length is not a hash this package produced.
	if len(salt) < minSaltLen || len(salt) > maxSaltLen {
		return false
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	if len(key) < minKeyLen || len(key) > maxKeyLen {
		return false
	}
	derived := argon2.IDKey([]byte(plain), salt, uint32(timeCost), uint32(memory), uint8(parallelism), uint32(len(key)))
	return subtle.ConstantTimeCompare(key, derived) == 1
}

// IsHashed reports whether the stored value is an argon2id hash.
func IsHashed(s string) bool {
	return strings.HasPrefix(s, hashPrefix)
}
