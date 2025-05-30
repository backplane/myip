.PHONY: build lint clean test

build: lint test myip

myip: main.go $(wildcard ./*/*.go)
	@echo '==> Building $@'
	go build -o "$@" -ldflags "\
        -X 'main.version=$$(git describe --tags --always --dirty)' \
        -X 'main.commit=$$(git rev-parse --short HEAD)' \
        -X 'main.date=$$(date -u +"%Y-%m-%dT%H:%M:%SZ")' \
        -X 'main.builtBy=make on $$(hostname)'"

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
