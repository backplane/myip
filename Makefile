.PHONY: build lint clean test

build: lint test myip

myip: main.go $(wildcard ./*/*.go)
	@echo '==> Building $@'
	go build -o "$@"

lint: main.go
	@echo '==> Linting'
	go fmt
	go vet
	staticcheck

test:
	@echo '==> Testing'
	go test -v clientip/*.go

clean:
	@echo '==> Cleaning'
	rm -rf -- myip
