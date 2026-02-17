PACKAGE=github.com/symbioticfi/relay
IMAGE_REPO ?= relay_sidecar
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

TAG ?=

ifeq ($(strip $(TAG)),)
	CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
	PSEUDO_VERSION := $(shell go list -f {{.Version}} -m ${PACKAGE}@${CURRENT_BRANCH} 2>/dev/null || echo "unspecified-$(CURRENT_BRANCH)")
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
ifeq ($(PUSH_LATEST), true)
	IMAGE_TAGS := ${IMAGE_TAGS} -t ${IMAGE_REPO}:latest
endif

.PHONY: local-setup
local-setup:
	cd e2e && \
	bash setup.sh && \
	cd temp-network && \
	docker compose up -d

.PHONY: clean-local-setup
clean-local-setup:
	if [ -d "e2e/temp-network" ]; then \
		docker compose --project-directory e2e/temp-network down; \
	fi

.PHONY: lint
lint: install-tools buf-lint go-lint

.PHONY: buf-lint
buf-lint:
	buf lint

.PHONY: go-lint
go-lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 -v run ./...

.PHONY: go-lint-fix
go-lint-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 -v run ./... --fix

.PHONY: generate
generate: install-tools generate-mocks generate-api-types generate-client-types generate-p2p-types generate-badger-types gen-abi generate-cli-docs

.PHONY: install-tools
install-tools:
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1
	go install go.uber.org/mock/mockgen@v0.6.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
	go install github.com/bufbuild/buf/cmd/buf@v1.59.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.7
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.7

.PHONY: generate-mocks
generate-mocks:
	go generate ./...

.PHONY: generate-api-types
generate-api-types:
	buf generate

.PHONY: generate-p2p-types
generate-p2p-types:
	buf generate --template=buf.p2p.gen.yaml

.PHONY: generate-badger-types
generate-badger-types:
	buf generate --template=buf.badger.gen.yaml

.PHONY: generate-client-types
generate-client-types:
	go run hack/codegen/generate-client-types.go

.PHONY: generate-cli-docs
generate-cli-docs:
	@echo "Generating CLI documentation..."
	go run hack/docgen/generate-cli-docs.go
	@echo "CLI documentation generated in docs/cli/"

.PHONY: unit-test
unit-test:
	go test $(shell go list ./... | grep -v '/e2e/') -v -covermode atomic -race -coverprofile=cover.out.tmp -coverpkg=$(shell go list ./... | grep -v '/e2e/' | tr '\n' ',')
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.tmp.txt # strip out generated files
	go tool cover -func coverage.tmp.txt > coverage.txt
	rm cover.out.tmp coverage.tmp.txt

.PHONY: e2e-test
e2e-test:
	cd e2e/tests && go test -v -timeout 40m

.PHONY: gen-abi
gen-abi:
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi symbiotic/client/evm/abi/ValSetDriver.abi.json \
		--type ValSetDriver \
		--pkg gen \
		--out symbiotic/client/evm/gen/valsetDriver.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi symbiotic/client/evm/abi/Settlement.abi.json \
		--type Settlement \
		--pkg gen \
		--out symbiotic/client/evm/gen/settlement.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi symbiotic/client/evm/abi/KeyRegistry.abi.json \
		--type KeyRegistry \
		--pkg gen \
		--out symbiotic/client/evm/gen/keyRegistry.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi symbiotic/client/evm/abi/VotingPowerProvider.abi.json \
		--type VotingPowerProvider \
		--pkg gen \
		--out symbiotic/client/evm/gen/votingPowerProvider.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi symbiotic/client/evm/abi/OperatorRegistry.abi.json \
		--type OperatorRegistry \
		--pkg gen \
		--out symbiotic/client/evm/gen/operatorRegistry.go

.PHONY: gen-abi-test
gen-abi-test:
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi e2e/tests/evm/abi/MockERC20.abi.json \
		--type MockERC20 \
		--pkg gen \
		--out e2e/tests/evm/gen/mockERC20.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi e2e/tests/evm/abi/IOptInService.abi.json \
		--type OptInService \
		--pkg gen \
		--out e2e/tests/evm/gen/optInService.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi e2e/tests/evm/abi/OpNetVaultAutoDeployLogic.abi.json \
		--type OpNetVaultAutoDeployLogic \
		--pkg gen \
		--out e2e/tests/evm/gen/opNetVaultAutoDeployLogic.go
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi e2e/tests/evm/abi/Vault.abi.json \
		--type Vault \
		--pkg gen \
		--out e2e/tests/evm/gen/vault.go

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
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'github.com/symbioticfi/relay/cmd/utils/root.Version=$(APP_VERSION)' -X 'github.com/symbioticfi/relay/cmd/utils/root.BuildTime=$(BUILD_TIME)'" -o relay_utils_$(OS)_$(ARCH) ./cmd/utils && \
		chmod a+x relay_utils_$(OS)_$(ARCH)

.PHONY: build-relay-sidecar
build-relay-sidecar:
	@if [ -z "$(OS)" ] || [ -z "$(ARCH)" ]; then \
		echo "Error: OS and ARCH parameters are required"; \
		echo "Usage: make build-relay-sidecar OS=<os> ARCH=<arch>"; \
		exit 1; \
	fi
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'github.com/symbioticfi/relay/cmd/relay/root.Version=$(APP_VERSION)' -X 'github.com/symbioticfi/relay/cmd/relay/root.BuildTime=$(BUILD_TIME)'" -o relay_sidecar_$(OS)_$(ARCH) ./cmd/relay && \
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
