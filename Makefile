BUF_VERSION:=1.35.1
SQLC_VERSION:=1.26.0

generate:
	rm -rf $$(pwd)/proto/*.pb.go $$(pwd)/users/db.go $$(pwd)/users/models.go $$(pwd)/users/querier.go $$(pwd)/users/users.sql.go
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) generate
	docker run -v $$(pwd)/users:/srv -w /srv sqlc/sqlc:$(SQLC_VERSION) generate
