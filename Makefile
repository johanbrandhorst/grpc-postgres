BUF_VERSION:=1.35.1
SQLC_VERSION:=1.26.0

generate:
	docker run -v $$(pwd):/srv -w /srv bufbuild/buf:$(BUF_VERSION) generate
	docker run -v $$(pwd)/users:/srv -w /srv sqlc/sqlc:$(SQLC_VERSION) generate
