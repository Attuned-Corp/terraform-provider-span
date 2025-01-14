
.PHONY: build
build:
	go build -v ./...

.PHONY: install
install: build
	go install -v ./...

lint:
	golangci-lint run --config ./.golangci.yml --fix

.PHONY: test
test:
	@TF_ACC=1 gotestsum --format pkgname -- --count=1 -cover -v ./...

ci-lint:
	@golangci-lint run --config ./.golangci.yml --timeout 5m && go mod tidy && git diff --exit-code -- go.mod go.sum
