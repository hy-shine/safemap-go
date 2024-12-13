package safemap

import "github.com/cespare/xxhash/v2"

func Hashstr(s string) uint64 {
	return xxhash.Sum64String(s)
}

func Hash(b []byte) uint64 {
	return xxhash.Sum64(b)
}
