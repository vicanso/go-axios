export GO111MODULE = on

.PHONY: default test test-cover bench lint dev

# for test
test:
	go test -race -cover ./...

bench:
	go test -bench=. -benchmem -run=^$$ ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

lint:
	golangci-lint run
