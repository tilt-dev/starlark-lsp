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
	goimports -w -l -local github.com/autokitteh/starlark-lsp cmd/ pkg/

install:
	go install ./cmd/autokitteh-starlark-lsp

builtins:
	go run ./hack/starlark-builtins.go > pkg/analysis/builtins.py


.PHONY: sysroot-pack sysroot-unpack release-dry-run release

PACKAGE_NAME          := github.com/autokitteh/autokitteh-starlark-lsp
GOLANG_CROSS_VERSION  ?= v1.21

DOCKER_RUN = docker run --rm -e CGO_ENABLED=1 -v /var/run/docker.sock:/var/run/docker.sock -v `pwd`:/go/src/$(PACKAGE_NAME) -v `pwd`/sysroot:/sysroot -w /go/src/$(PACKAGE_NAME)
GORELEASER_IMAGE = ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION}

release-dry-run:
	@$(DOCKER_RUN) $(GORELEASER_IMAGE) --clean --skip=validate --skip=publish

release:
	@$(DOCKER_RUN) --env-file .release-env $(GORELEASER_IMAGE) release --clean
