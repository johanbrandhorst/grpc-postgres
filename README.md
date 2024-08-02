# grpc-postgres
An example repo of how I like to use postgres with gRPC
```
  +------------------------+
  |   ____  ______  ___    |
  |  /    )/      \/   \   |
  | (     / __    _\    )  |
  |  \    (/ o)  ( o)   )  |
  |   \_  (_  )   \ )  /   |
  |     \  /\_/    \)_/    |
  |      \/  //|  |\\      |
  |          v |  | v      |
  |            \__/        |
  |                        |
  +------------------------+
```
Source: [Charles Clavadetscher](https://www.postgresql.org/message-id/57386570.8090703%40swisspug.org)

## Usage

First, start a postgres container:

```bash
$ docker run --rm -d --name postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=mypass -e POSTGRES_DB=postgres postgres:13
d1a2eb0fb44da9c3488184f5296da28d1c7f88bd32bd4ec81fc254f006886b03
```

Start the server:

```bash
$ POSTGRES_URL=postgresql://postgres:mypass@localhost:5432/postgres?sslmode=disable go run main.go
...
{"level":"info","msg":"Serving gRPC on [::]:8080"}
{"level":"info","msg":"Serving Web UI on http://0.0.0.0:8080"}
```

Navigate to http://0.0.0.0:8080 to see the auto-generated web UI for the
service, courtesy of gRPC reflection and
[github.com/fullstorydev/grpcui](https://github.com/fullstorydev/grpcui/)!

![gRPCUI](./grpcui.png)

## Usage with Cockroach DB

The application also supports talking to a Cockroach DB instance, by using the `cockroachdb` scheme:

```
cockroachdb://user:password@host:5432/defaultdb
```

The scheme is used when performing the database migrations, as the behaviour changes based on the database.

## Developing

### Requirements

* `go` > 1.21

### Making changes

After making any changes to the proto file or the migrations, make sure to
regenerate the files:

```bash
$ make generate
```

If you want to change the schema of the database, add another migration file.
Use the same naming format, with the new file names starting `2_`. Make sure to
run `make generate` and increment the migration version in `users/helpers.go`.
