lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 -v run ./...

install-mocks:
	go install go.uber.org/mock/mockgen@latest

gen-mocks:
	go generate ./...

unit-test:
	go test ./... -v -covermode atomic -race -coverprofile=cover.out.tmp  -coverpkg=./...
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.tmp.txt # strip out generated files
	go tool cover -func coverage.tmp.txt > coverage.txt
	rm cover.out.tmp coverage.tmp.txt

build-for-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o middleware-offchain ./ && chmod a+x middleware-offchain


gen-abi:
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi internal/client/symbiotic/Master.abi.json \
		--type Master \
		--pkg gen \
		--out internal/client/symbiotic/gen/master.go


build-generate-genesis-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o generate_genesis_linux_amd64 ./cmd/generate-genesis && \
		chmod a+x generate_genesis_linux_amd64

build-generate-genesis-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o generate_genesis_darwin_arm64 ./cmd/generate-genesis && \
		chmod a+x generate_genesis_darwin_arm64

build-middleware-offchain-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o middleware_offchain_darwin_arm64 ./cmd/middleware-offchain && \
		chmod a+x middleware_offchain_darwin_arm64

build-docker:
	docker build -t middleware-offchain .

fix-goimports:
	go run golang.org/x/tools/cmd/goimports@latest -w .
