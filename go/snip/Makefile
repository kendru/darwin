.PHONY: build
build:
	go build .

.PHONY: clean
clean:
	rm -rf snip data

.PHONY: test
test:
	go clean -testcache
	go test ./...

.PHONY: run
run:
	go run main.go
