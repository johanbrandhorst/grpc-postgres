# grpc-postgres
An example repo of how I like to use postgres with gRPC

## Usage

First, start a postgres container:

```bash
$ docker run --rm -d --name postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mypass -e POSTGRES_DB=postgres postgres
d1a2eb0fb44da9c3488184f5296da28d1c7f88bd32bd4ec81fc254f006886b03
```

Start the server:

```bash
$ go run main.go --postgres-url psql://postgres:mypass@localhost:5432/postgres
...
INFO[May 21 22:51:37.091] Serving gRPC on [::]:10000
INFO[May 21 22:51:37.156] Serving Web UI on https://localhost:10000
```

Navigate to http://localhost:10000 to see the auto-generated web UI for the
service, courtesy of gRPC reflection and
[github.com/fullstorydev/grpcui](github.com/fullstorydev/grpcui)!

## Developing

### Requirements

* `protoc` > 3.0.0
* `go` > 1.11

Run `make install` to install the generation dependencies:

```bash
$ make install
go get \
        github.com/golang/protobuf/protoc-gen-go \
        github.com/jteeuwen/go-bindata/go-bindata \
        github.com/golang/mock/mockgen
go: finding github.com/jteeuwen/go-bindata/go-bindata latest
go: finding github.com/golang/mock/mockgen latest
go: finding github.com/golang/protobuf/protoc-gen-go latest
```

### Making changes

After making any changes to the proto file or the migrations, make sure to
regenerate the files:

```bash
$ make generate
protoc -I proto --go_out=plugins=grpc,paths=source_relative:./proto ./proto/users.proto
go-bindata -pkg migrations -ignore bindata -prefix ./users/migrations/ -o ./users/migrations/bindata.go ./users/migrations
mockgen -destination ./users/mocks_test.go -package users_test github.com/johanbrandhorst/grpc-postgres/proto UserService_ListUsersServer
```

If you want to change the schema of the database, add another migration file.
Use the same naming format, with the new file names starting `2_`. Make sure to
run `make generate` and increment the migration version in `users/helpers.go`.
