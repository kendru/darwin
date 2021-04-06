.PHONY: build
build:
	go build .

.PHONY: clean
clean:
	rm eavt

.PHONY: test
test:
	go clean -testcache
	go test ./...

.PHONY: run
run:
	go run main.go
