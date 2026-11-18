package rands

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// hexAlphabet is the lowercase hexadecimal alphabet used by Hex.
const hexAlphabet = "0123456789abcdef"

// cryptoIntN returns a uniformly distributed integer in [0, n) using the
// operating system CSPRNG. It panics only when the system entropy source is
// unavailable, which is unrecoverable for a library that promises crypto-safe
// output; callers must not faintly "fall back" to math/rand.
func cryptoIntN(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("rands: crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

// Crypto generates a cryptographically secure random alphanumeric string of the
// specified length. Use it for secrets, salts, token identifiers and generated
// passwords — anything an attacker must not be able to predict.
//
// For non-security purposes (jitter, sampling, test data) prefer CharFast, which
// is faster but backed by math/rand/v2 and therefore predictable.
func Crypto(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[cryptoIntN(len(alphabet))]
	}
	return string(b)
}

// Hex generates a cryptographically secure random lowercase hex string with
// exactly length characters (length bytes are drawn from crypto/rand, so length
// must be even).
func Hex(length int) string {
	if length <= 0 {
		return ""
	}
	// +1 guards against an odd length rounding down to nothing.
	buf := make([]byte, (length+1)/2)
	if _, err := rand.Read(buf); err != nil {
		panic("rands: crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(buf)[:length]
}

// CryptoDigits generates a cryptographically secure random decimal digit string
// of the specified length. Leading zeros are preserved.
func CryptoDigits(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = byte('0' + cryptoIntN(10))
	}
	return string(b)
}

// CryptoPick returns a random element from a slice using the CSPRNG. It returns
// the zero value for an empty slice.
func CryptoPick[T any](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}
	return list[cryptoIntN(len(list))]
}
