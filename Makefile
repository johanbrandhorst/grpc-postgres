install:
	go get \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/tmthrgd/go-bindata/go-bindata \
		github.com/golang/mock/mockgen \
		github.com/bufbuild/buf/cmd/buf \
		github.com/kyleconroy/sqlc/cmd/sqlc

generate:
	buf protoc -I proto --go_out=paths=source_relative:./proto --go-grpc_out=paths=source_relative:./proto ./proto/users.proto
	go-bindata -pkg migrations -ignore bindata -nometadata -prefix users/migrations/ -o ./users/migrations/bindata.go ./users/migrations
	mockgen -destination ./users/mocks_test.go -package users_test github.com/johanbrandhorst/grpc-postgres/proto UserService_ListUsersServer
	cd users && sqlc generate 