package safemap

import "github.com/cespare/xxhash/v2"

// Hashstr computes the xxHash of a string.
// It returns a 64-bit hash value suitable for use as a map key hash.
func Hashstr(s string) uint64 {
	return xxhash.Sum64String(s)
}

// Hash computes the xxHash of a byte slice.
// It returns a 64-bit hash value suitable for use as a map key hash.
func Hash(b []byte) uint64 {
	return xxhash.Sum64(b)
}
