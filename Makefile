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


build-generate-genesis-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o generate_genesis_linux_amd64 ./cmd/generate-genesis && \
		chmod a+x generate_genesis_linux_amd64

build-generate-genesis-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o generate_genesis_darwin_arm64 ./cmd/generate-genesis && \
		chmod a+x generate_genesis_darwin_arm64

build-symbiotic-relay-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o symbiotic_relay_linux_amd64 ./cmd/middleware-offchain && \
		chmod a+x symbiotic_relay_linux_amd64

build-symbiotic-relay-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o symbiotic_relay_darwin_arm64 ./cmd/middleware-offchain && \
		chmod a+x symbiotic_relay_darwin_arm64

build-docker:
	docker build -t middleware-offchain .

fix-goimports:
	go run golang.org/x/tools/cmd/goimports@latest -w .
