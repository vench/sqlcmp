NAME=sqlcmp
BINFILE=sqlcmp
PWD=$(shell pwd)

GOLANGCI_LINT=$(shell which golangci-lint)
GO_IMPORT_LINT=$(shell which go-import-lint)


precommit: linter test build

linter:
	@test ! -x "$(GOLANGCI_LINT)" || $(GOLANGCI_LINT) run -c ./golangci.yml  $(PWD)/...;

test:
	@mkdir -p ./.coverage
	go test -race -cover -covermode=atomic -coverprofile=.coverage/cover.out $(PWD)/...

build:
	@mkdir -p ./bin
	go build -o ./bin/$(BINFILE)





