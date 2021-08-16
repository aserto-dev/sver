SHELL 	   := $(shell which bash)

## BOF define block

# This is required in order to reference private repos
# this works in tandem with the following git configuration, defined below
# git config --global url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE=github.com/aserto-dev

GIT_ON_SSH 				:= $(git config --global url."git@github.com:".insteadOf "https://github.com/")
SSH_PRIVATE_KEY_FILE 	?= $(HOME)/.ssh/id_rsa
SSH_PRIVATE_KEY 		?= $(file < $(SSH_PRIVATE_KEY_FILE))

GOARCH     ?= amd64
GOOS       := $(shell go env GOOS)
CGO_ENABLED:=0
LDBASE     := main
LDFLAGS    := -ldflags "-X ${LDBASE}.ver=${VERSION} -X ${LDBASE}.date=${DATE} -X ${LDBASE}.commit=${COMMIT}"
GOPATH     := $(shell go env GOPATH)

VERSION    ?= v$(shell sver 2>/dev/null)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE       ?= $(shell date "+%FT%T%z")

export VERSION
export SSH_PRIVATE_KEY
export COMMIT

TARGET     := sver
ROOT_DIR   ?= $(shell git rev-parse --show-toplevel)
BIN_DIR    := ${ROOT_DIR}/bin
SRC_DIR    := ${ROOT_DIR}
DIST_DIR   := ${ROOT_DIR}/dist
BIN_FILE   := ${BIN_DIR}/${GOOS}-${GOARCH}/${TARGET}$(if $(findstring ${GOOS},windows),".exe","")

${BIN_DIR}:
	@echo -e "${ATTN_COLOR}==> create BIN_DIR ${BIN_DIR} ${NO_COLOR}"
	@mkdir -p ${BIN_DIR}

TESTER     := ${GOPATH}/bin/gotestsum
${TESTER}:
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@go install gotest.tools/gotestsum@v1.7.0

LINTER	   := ${GOPATH}/bin/golangci-lint
${LINTER}:
	@echo -e "${ATTN_COLOR}==> $@  ${NO_COLOR}"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1

RELEASER   := ${GOPATH}/bin/goreleaser
${RELEASER}:
	@echo -e "${ATTN_COLOR}==> $@  ${NO_COLOR}"
	@go install github.com/goreleaser/goreleaser@v0.174.2

.PHONY: all
all: deps build test lint

.PHONY: deps
deps:
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@go install .

.PHONY: build
build: ${BIN_DIR} deps
	@echo -e "${ATTN_COLOR}==> GOOS=${GOOS} GOARCH=${GOARCH} VERSION=${VERSION} COMMIT=${COMMIT} DATE=${DATE} ${NO_COLOR}"
	@GOOS=${GOOS} GOARCH=${GOARCH} GO111MODULE=on go build ${LDFLAGS} -o ${BIN_FILE} ${SRC_DIR}

ifneq (${GOOS},windows)
	@chmod +x ${BIN_FILE}
endif

.PHONY: test 
test: ${TESTER} build
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@${TESTER} --format short-verbose -- -coverprofile=cover.out -coverpkg=./... -count=1 -timeout 90s -v ${ROOT_DIR}/...

.PHONY: lint
lint: ${LINTER} build
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@${LINTER} run
	@echo -e "${NO_COLOR}\c"

.PHONY: release
release: ${RELEASER} build
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@${RELEASER} release --skip-publish --rm-dist --snapshot --config .goreleaser.yml

.PHONY: publish
publish: ${RELEASER} build
ifndef HOMEBREW_TAP
	$(error HOMEBREW_TAP environment variable is undefined)
endif
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@${RELEASER} release --config .goreleaser.yml --rm-dist

.PHONY: clean
clean:
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@rm -rf ${BIN_DIR}
	@rm -rf $(DIST_DIR)

.PHONY: docker-image
docker-image:
	@echo -e "${ATTN_COLOR}==> $@ ${NO_COLOR}"
	@docker build . --build-arg VERSION --build-arg SSH_PRIVATE_KEY --build-arg COMMIT -t aserto/sver:$(VERSION)
