BUF_VERSION:=1.0.0-rc6
SQLC_VERSION:=1.10.0

generate:
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) generate
	docker run -v $$(pwd)/users:/srv -w /srv kjconroy/sqlc:$(SQLC_VERSION) generate
