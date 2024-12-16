PONY: bench test

test:

benchAll:
	@echo "Run: make benchAll"
	go test -benchmem -bench .

benchConcurrency:
	@echo "Run: make benchConcurrency"
	go test -benchmem -bench=_Concurent.* .

benchSingle:
	@echo "Run: make benchSingle"
	go test -benchmem -bench=_Single* .