PACKAGE=github.com/symbioticfi/relay
IMAGE_REPO ?= relay_sidecar
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

TAG ?=

ifeq ($(strip $(TAG)),)
	CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
	PSEUDO_VERSION := $(shell go list -f {{.Version}} -m ${PACKAGE}@${CURRENT_BRANCH})
	# Trim the `v` prefix from golang pseudo version as the TAG if not set
	FINAL_TAG := $(shell echo $(PSEUDO_VERSION) | sed 's/^v//' | sed 's/-0\./-/')
else
	# If TAG was explicitly passed, strip the v prefix
	TAG_ORIGINAL := $(TAG)
	FINAL_TAG := $(shell echo '$(TAG_ORIGINAL)' | sed 's/^v//')
endif

# add v prefix for APP_VERSION
APP_VERSION := v$(FINAL_TAG)

# create image tags without v prefix
IMAGE_TAGS := -t ${IMAGE_REPO}:${FINAL_TAG}

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1 -v run ./...

.PHONY: install-mocks
install-mocks:
	go install go.uber.org/mock/mockgen@latest

.PHONY: gen-mocks
gen-mocks:
	go generate ./...

.PHONY: gen-api
gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@v1.14.0 -v -clean  -package api -target internal/gen/api api/swagger.yaml

.PHONY: unit-test
unit-test:
	go test ./... -v -covermode atomic -race -coverprofile=cover.out.tmp  -coverpkg=./...
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.tmp.txt # strip out generated files
	go tool cover -func coverage.tmp.txt > coverage.txt
	rm cover.out.tmp coverage.tmp.txt

.PHONY: gen-abi
gen-abi:
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi core/client/evm/abi/IValSetDriver.abi.json \
		--type IValSetDriver \
		--pkg gen \
		--out core/client/evm/gen/valsetDriver.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi core/client/evm/abi/ISettlement.abi.json \
		--type ISettlement \
		--pkg gen \
		--out core/client/evm/gen/settlement.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi core/client/evm/abi/IKeyRegistry.abi.json \
		--type IKeyRegistry \
		--pkg gen \
		--out core/client/evm/gen/keyRegistry.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi core/client/evm/abi/IVotingPowerProvider.abi.json \
		--type IVotingPowerProvider \
		--pkg gen \
		--out core/client/evm/gen/votingPowerProvider.go

# Generic build target that takes OS and architecture as parameters
# Usage: make build-relay-utils OS=linux ARCH=amd64
# Usage: make build-relay-sidecar OS=darwin ARCH=arm64
.PHONY: build-relay-utils
build-relay-utils:
	@if [ -z "$(OS)" ] || [ -z "$(ARCH)" ]; then \
		echo "Error: OS and ARCH parameters are required"; \
		echo "Usage: make build-relay-utils OS=<os> ARCH=<arch>"; \
		exit 1; \
	fi
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_utils_$(OS)_$(ARCH) ./cmd/utils && \
		chmod a+x relay_utils_$(OS)_$(ARCH)

.PHONY: build-relay-sidecar
build-relay-sidecar:
	@if [ -z "$(OS)" ] || [ -z "$(ARCH)" ]; then \
		echo "Error: OS and ARCH parameters are required"; \
		echo "Usage: make build-relay-sidecar OS=<os> ARCH=<arch>"; \
		exit 1; \
	fi
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_sidecar_$(OS)_$(ARCH) ./cmd/relay && \
		chmod a+x relay_sidecar_$(OS)_$(ARCH)

# Legacy targets for backward compatibility
.PHONY: build-relay-utils-linux
build-relay-utils-linux:
	$(MAKE) build-relay-utils OS=linux ARCH=amd64

.PHONY: build-relay-utils-darwin
build-relay-utils-darwin:
	$(MAKE) build-relay-utils OS=darwin ARCH=arm64

.PHONY: build-relay-sidecar-linux
build-relay-sidecar-linux:
	$(MAKE) build-relay-sidecar OS=linux ARCH=amd64

.PHONY: build-relay-sidecar-darwin
build-relay-sidecar-darwin:
	$(MAKE) build-relay-sidecar OS=darwin ARCH=arm64

.PHONY: image
image:
ifeq ($(PUSH_IMAGE), true)
	@docker buildx build --push --platform=linux/amd64,linux/arm64 . ${IMAGE_TAGS} --build-arg APP_VERSION=$(APP_VERSION) --build-arg BUILD_TIME=$(BUILD_TIME)
	# https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
	echo "image=${IMAGE_REPO}:${FINAL_TAG}" >> $$GITHUB_OUTPUT
else
	@DOCKER_BUILDKIT=1 docker build . ${IMAGE_TAGS} --build-arg APP_VERSION=$(APP_VERSION) --build-arg BUILD_TIME=$(BUILD_TIME)
endif

.PHONY: fix-goimports
fix-goimports:
	go run golang.org/x/tools/cmd/goimports@latest -w .
