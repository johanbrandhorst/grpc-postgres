BUF_VERSION:=0.40.0
SQLC_VERSION:=1.8.0

install:
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc
	curl -sSL \
		"https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-Linux-x86_64" \
		-o "$(shell go env GOPATH)/bin/buf"
	chmod +x "$(shell go env GOPATH)/bin/buf"
	curl -sSL \
		"https://github.com/kyleconroy/sqlc/releases/download/v$(SQLC_VERSION)/sqlc-v$(SQLC_VERSION)-linux-amd64.zip" \
		-o sqlc.zip
	unzip -o -d "$(shell go env GOPATH)/bin/" sqlc.zip
	chmod +x "$(shell go env GOPATH)/bin/sqlc"
	rm sqlc.zip

generate:
	buf generate
	cd users && sqlc generate
