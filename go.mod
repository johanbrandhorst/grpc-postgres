module github.com/johanbrandhorst/grpc-postgres

go 1.13

require (
	github.com/Masterminds/squirrel v1.4.0
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/antlr/antlr4 v0.0.0-20200911155845-1ac3593fc4e1 // indirect
	github.com/bufbuild/buf v0.23.0
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/fullstorydev/grpcui v1.0.0
	github.com/fullstorydev/grpcurl v1.7.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.12.2
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.2
	github.com/google/gxui v0.0.0-20151028112939-f85e0a97b3a4 // indirect
	github.com/jackc/pgproto3/v2 v2.0.4 // indirect
	github.com/jackc/pgtype v1.4.2
	github.com/jackc/pgx/v4 v4.8.1
	github.com/kr/text v0.2.0 // indirect
	github.com/kyleconroy/sqlc v1.5.0
	github.com/lib/pq v1.8.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/ory/dockertest/v3 v3.6.0
	github.com/pingcap/log v0.0.0-20200828042413-fce0951f1463 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/tmthrgd/go-bindata v0.0.0-20190904063317-a4b65675e0fb
	github.com/tmthrgd/go-rand v0.0.0-20190904060720-34764beea44d // indirect
	github.com/zach-klippenstein/goregen v0.0.0-20160303162051-795b5e3961ea // indirect
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	golang.org/x/tools v0.0.0-20200911193555-6422fca01df9 // indirect
	google.golang.org/grpc v1.32.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v0.0.0-20200910201057-6591123024b3
	google.golang.org/grpc/examples v0.0.0-20200910201057-6591123024b3 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	vitess.io/vitess v0.7.0 // indirect
)

// https://github.com/kyleconroy/sqlc/issues/654
replace github.com/pingcap/parser => github.com/kyleconroy/parser v0.0.0-20200819185651-2caf0f596c0c
