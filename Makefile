install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/jteeuwen/go-bindata/go-bindata

generate:
	protoc -I proto --go_out=plugins=grpc,paths=source_relative:./proto ./proto/users.proto
	go-bindata -pkg migrations -ignore bindata -prefix ./users/migrations/ -o ./users/migrations/bindata.go ./users/migrations
