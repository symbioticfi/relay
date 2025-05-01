lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2 -v run ./...

unit-test::
	go test ./... -v -race -covermode atomic -coverprofile=cover.out.tmp  -coverpkg=./...
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.txt # strip out generated files
	go tool cover -func coverage.txt
	rm cover.out.tmp coverage.txt

build-for-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static'" -o middleware-offchain ./ && chmod a+x middleware-offchain


gen-abi:
	go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
		--abi internal/client/eth/abi/IMasterConfigManager.abi.json \
		--pkg eth \
		--type IMasterConfigManager \
		--out internal/client/eth/abi/IMasterConfigManager.go

#gen-abi:
#	@for file in internal/client/eth/abi*.abi.json; do \
#		filename=$$(basename $$file .abi.json); \
#		echo "Generating $$filename..."; \
#		go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
#			--abi $$file \
#			--pkg eth \
#			--type $$filename \
#			--out internal/client/eth/$$filename.go; \
#	done