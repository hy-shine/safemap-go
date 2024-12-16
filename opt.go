package safemap

type options[K comparable] struct {
	bucketTotal int
	hashFunc    func(K) uint64
}

type OptFunc[K comparable] func(*options[K])

// WithBuckets sets safemap buckets capacity
// bit: 0-8
func WithBuckets[K comparable](bit uint8) OptFunc[K] {
	return func(o *options[K]) {
		if bit > 8 {
			bit = 8
		}
		o.bucketTotal = int(1 << bit)
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

	if opt.bucketTotal == 0 {
		opt.bucketTotal = defaultBucketCount
	}
	if opt.bucketTotal > maxBucketCount {
		opt.bucketTotal = maxBucketCount
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
