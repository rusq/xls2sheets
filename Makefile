SHELL=/bin/sh

dist:
	goreleaser check
	goreleaser release --snapshot --clean
.PHONY: dist

