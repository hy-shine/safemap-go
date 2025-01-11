package safemap

type options[K comparable] struct {
	bucketTotal int
	hashFunc    func(K) uint64
}

type OptFunc[K comparable] func(*options[K])

// WithBuckets sets safemap buckets capacity
func WithBuckets[K comparable](mask uint8) OptFunc[K] {
	return func(o *options[K]) {
		if 1<<mask > maxBucketCount {
			o.bucketTotal = maxBucketCount
		} else {
			o.bucketTotal = int(1 << mask)
		}
	}
}

// WithHashFunc sets hash function for key.
func WithHashFunc[K comparable](fn func(K) uint64) OptFunc[K] {
	return func(o *options[K]) {
		o.hashFunc = fn
	}
}

func loadOpts[K comparable](opts ...OptFunc[K]) (*options[K], error) {
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

func HashStrKeyFunc() OptFunc[string] {
	return func(o *options[string]) {
		o.hashFunc = Hashstr
	}
}
