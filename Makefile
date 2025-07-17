lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1 -v run ./...

install-mocks:
	go install go.uber.org/mock/mockgen@latest

gen-mocks:
	go generate ./...

gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@v1.14.0 -v -clean  -package api -target internal/gen/api api/swagger.yaml

unit-test:
	go test ./... -v -covermode atomic -race -coverprofile=cover.out.tmp  -coverpkg=./...
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.tmp.txt # strip out generated files
	go tool cover -func coverage.tmp.txt > coverage.txt
	rm cover.out.tmp coverage.tmp.txt

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

APP_VERSION ?= dev
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build-relay-utils-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_utils_linux_amd64 ./cmd/utils && \
		chmod a+x relay_utils_linux_amd64

build-relay-utils-darwin:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_utils_darwin_arm64 ./cmd/utils && \
		chmod a+x relay_utils_darwin_arm64

build-relay-sidecar-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_sidecar_linux_amd64 ./cmd/relay_sidecar && \
		chmod a+x relay_sidecar_linux_amd64

build-relay-sidecar-darwin:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=$(APP_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" -o relay_sidecar_darwin_arm64 ./cmd/relay_sidecar && \
		chmod a+x relay_sidecar_darwin_arm64

build-docker:
	docker build -t relay_sidecar .

fix-goimports:
	go run golang.org/x/tools/cmd/goimports@latest -w .
