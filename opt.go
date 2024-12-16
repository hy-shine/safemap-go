package safemap

type options[K comparable] struct {
	lockCap  int
	hashFunc func(K) uint64
}

type OptFunc[K comparable] func(*options[K])

// WithCap sets map capacity
// bit: 0-8
func WithCap[K comparable](bit uint8) OptFunc[K] {
	return func(o *options[K]) {
		if bit > 8 {
			bit = 8
		}
		o.lockCap = int(1 << bit)
	}
}

// WithHashFunc sets hash function for key.
func WithHashFunc[K comparable](fn func(K) uint64) OptFunc[K] {
	return func(o *options[K]) {
		o.hashFunc = fn
	}
}

func loadOptfuns[K comparable](opts ...OptFunc[K]) (*options[K], error) {
	opt := &options[K]{}
	for i := range opts {
		opts[i](opt)
	}

	if opt.lockCap == 0 {
		opt.lockCap = defaultLockCount
	}
	if opt.lockCap > maxLockCount {
		opt.lockCap = maxLockCount
	}
	if opt.hashFunc == nil {
		return nil, ErrMissingHashFunc
	}

	return opt, nil
}

func HashstrKeyFunc() OptFunc[string] {
	return func(o *options[string]) {
		o.hashFunc = Hashstr
	}
}
