
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
