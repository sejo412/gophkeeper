MODULE := github.com/sejo412/gophkeeper/internal/config
BUILD_VERSION ?= 0.0.0-rc1
BUILD_COMMIT ?= $$(git rev-parse HEAD)
BUILD_DATE ?= $$(date -R)

.PHONY: all
all: clean proto server client

.PHONY: server
server:
	go build -race -ldflags \
		"-X '$(MODULE).BuildVersion=$(BUILD_VERSION)'\
		-X '$(MODULE).BuildCommit=$(BUILD_COMMIT)'\
		-X '$(MODULE).BuildDate=$(BUILD_DATE)'"\
		-o ./bin/server ./cmd/server/

.PHONY: client
client:
	go build -race -ldflags \
		"-X '$(MODULE).BuildVersion=$(BUILD_VERSION)'\
		-X '$(MODULE).BuildCommit=$(BUILD_COMMIT)'\
		-X '$(MODULE).BuildDate=$(BUILD_DATE)'"\
		-o ./bin/client ./cmd/client/

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

.PHONY: clean
clean:
	rm -f ./bin/{server,client}
