package safemap

type opt[K comparable] struct {
	lock     int
	hashFunc func(K) uint64
}

type OptFunc[K comparable] func(*opt[K])

func WithCap[K comparable](bit uint8) OptFunc[K] {
	return func(o *opt[K]) {
		if bit > 8 {
			bit = 8
		}
		o.lock = int(1 << bit)
	}
}

func WithHashFn[K comparable](fn func(K) uint64) OptFunc[K] {
	return func(o *opt[K]) {
		o.hashFunc = fn
	}
}

func loadOptfuns[K comparable](opts ...OptFunc[K]) (*opt[K], error) {
	_opt := &opt[K]{}
	for i := range opts {
		opts[i](_opt)
	}

	if _opt.lock == 0 {
		_opt.lock = defaultLockCount
	}
	if _opt.lock > maxLockCount {
		_opt.lock = maxLockCount
	}
	if _opt.hashFunc == nil {
		return nil, ErrMissingHashFunc
	}

	return _opt, nil
}
