BUF_VERSION:=1.17.0
SQLC_VERSION:=1.18.0

generate:
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) generate
	docker run -v $$(pwd)/users:/srv -w /srv kjconroy/sqlc:$(SQLC_VERSION) generate
