lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 -v run ./...

unit-test::
	go test ./... -v -race -covermode atomic -coverprofile=cover.out.tmp  -coverpkg=./...
	cat cover.out.tmp | grep -v "gen"  | grep -v "mocks" > coverage.txt # strip out generated files
	go tool cover -func coverage.txt
	rm cover.out.tmp coverage.txt
