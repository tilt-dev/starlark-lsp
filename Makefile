.PHONY: test lint tidy fmt install

test:
ifneq ($(CIRCLECI),true)
	gotestsum -- ./... -timeout 30s
else
	mkdir -p test-results
	gotestsum --format standard-quiet --junitfile test-results/unit-tests.xml -- ./... -timeout 30s
endif

lint:
ifneq ($(CIRCLECI),true)
	golangci-lint run
else
	mkdir -p test-results
	golangci-lint run --out-format="junit-xml" --new-from-rev="HEAD~" > ./test-results/lint.xml
endif

tidy:
	go mod tidy

fmt:
	goimports -w -l -local github.com/tilt-dev/starlark-lsp cmd/ pkg/

install:
	go install ./cmd/starlark-lsp

builtins:
	go run ./hack/starlark-builtins.go > pkg/analysis/builtins.py
