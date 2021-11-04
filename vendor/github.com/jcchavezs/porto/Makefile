.DEFAULT_GOAL := test

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run ./... && echo "Done."

.PHONY: deps
deps:
	@go get -v -t -d ./...

.PHONY: fmt
fmt:
	gofmt -w -s ./

.PHONY: install-tools
install-tools: ## Install all the dependencies under the tools module
	$(MAKE) -C ./tools install