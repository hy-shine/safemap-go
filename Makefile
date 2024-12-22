PONY: test benchAll benchConcurrent benchSingle

test:
	@echo "Run: make test"
	go test -v ./...

benchAll:
	@echo "Run: make benchAll"
	go test -benchmem -bench .

benchConcurrent:
	@echo "Run: make benchConcurrent"
	go test -benchmem -bench=^Benchmark_Concurrent.* .

benchSingle:
	@echo "Run: make benchSingle"
	go test -benchmem -bench=^Benchmark_Single.* .

benchBucket:
	@echo "Run: make benchBucket"
	go test -benchmem -bench=^Benchmark_Bucket.* .