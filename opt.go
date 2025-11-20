package safemap

type options[K comparable] struct {
	bucketTotal int
	hashFunc    func(K) uint64
}

type OptFunc[K comparable] func(*options[K])

// WithBuckets sets the number of buckets to 1<<mask.
// For example, WithBuckets(5) creates 32 buckets (1<<5).
// If mask is 0, the default bucket count (32) is used.
// If 1<<mask exceeds maxBucketCount (1024), it is capped at the maximum.
func WithBuckets[K comparable](mask uint8) OptFunc[K] {
	return func(o *options[K]) {
		if mask == 0 {
			// Use default bucket count
			o.bucketTotal = defaultBucketCount
		} else if 1<<mask > maxBucketCount {
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

func buildOptions[K comparable](opts ...OptFunc[K]) (*options[K], error) {
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
