install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/tmthrgd/go-bindata/go-bindata \
		github.com/golang/mock/mockgen

generate:
	protoc -I proto --go_out=plugins=grpc,paths=source_relative:./proto ./proto/users.proto
	go-bindata -pkg migrations -ignore bindata -nometadata -prefix users/migrations/ -o ./users/migrations/bindata.go ./users/migrations
	mockgen -destination ./users/mocks_test.go -package users_test github.com/johanbrandhorst/grpc-postgres/proto UserService_ListUsersServer
