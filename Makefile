
.PHONY: statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

.PHONY: fieldalignment-diff
fieldalignment-diff:
	fieldalignment -fix -diff ./...

.PHONY: fieldalignment-fix
fieldalignment-fix:
	fieldalignment -fix ./...

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

.PHONY: cover
cover:
	#go test ./... -coverprofile=./coverage.out -covermode=atomic -coverpkg=./...
	go test -v -tags integration -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@rm -f coverage.out
