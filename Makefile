SHELL 	   := $(shell which bash)

## BOF define block

# This is required in order to reference private repos
# this works in tandem with the following git configuration, defined below
# git config --global url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE=github.com/aserto-dev
GIT_ON_SSH := $(git config --global url."git@github.com:".insteadOf "https://github.com/")

ROOT_DIR   ?= $(shell git rev-parse --show-toplevel)
BIN_DIR    := $(ROOT_DIR)/bin
SRC_DIR    := $(ROOT_DIR)/
BINARIES   := calc-version

COMMIT     ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION    ?= v`calc-version`
DATE       ?= `date "+%FT%T%z"`

LDBASE     := main
DEV_LDFLAGS:= -ldflags "-X $(LDBASE).ver=${VERSION} -X $(LDBASE).date=${DATE} -X $(LDBASE).commit=${COMMIT}"
GOARCH     ?= amd64
GOOS       := $(shell go env GOOS)
LINTER     := $(BIN_DIR)/golangci-lint
LINTVERSION:= v1.32.2

$(LINTER):
	@echo -e "$(ATTN_COLOR)==> get $@  $(NO_COLOR)"
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s $(LINTVERSION)

.PHONY: lint
lint: $(LINTER)
	$(LINTER) run

.PHONY: deps
deps:
	@go install -i github.com/aserto-dev/calc-version

.PHONY: dobuild
dobuild:
	@echo $(DEV_LDFLAGS)
	@GOOS=$(P) GOARCH=$(GOARCH) GO111MODULE=on go build $(DEV_LDFLAGS) -o $(T)/$(P)-$(GOARCH)/$(B)$(if $(findstring $(P),windows),".exe","") $(SRC_DIR)
ifneq ($(P),windows)
	@chmod +x $(T)/$(P)-$(GOARCH)/$(B)
endif

.PHONY: build
build: $(BIN_DIR)
	@for b in ${BINARIES};	\
	do	\
		$(MAKE) dobuild B=$${b} P=${GOOS} T=${BIN_DIR};	\
	done
