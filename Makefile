# Author:  Niels A.D.
# Project: goop (https://github.com/nielsAD/goop)
# License: Mozilla Public License, v2.0

GOW3=vendor/github.com/nielsAD/gowarcraft3
VENDOR=$(GOW3)/vendor/bncsutil/build/libbncsutil_static.a

GO_FLAGS=
GOTEST_FLAGS=-cover -cpu=1,2,4 -timeout=2m

GO=go
GOFMT=gofmt
GOLINT=$(shell $(GO) env GOPATH)/bin/golint

DIR_BIN=bin
DIR_PRE=github.com/nielsAD/goop

PKG:=$(shell $(GO) list ./...)
DIR:=$(subst $(DIR_PRE),.,$(PKG))

ARCH:=$(shell $(GO) env GOARCH)
ifeq ($(ARCH),amd64)
	TEST_RACE=1
endif

ifeq ($(TEST_RACE),1)
	GOTEST_FLAGS+= -race
endif

.PHONY: all release check test fmt lint vet list clean

all: test release

$(DIR_BIN):
	mkdir -p $@

$(PKG): $(VENDOR)
	$(GO) build $@

$(GOW3)/vendor/%:
	$(MAKE) -C $(GOW3)/vendor $(subst $(GOW3)/vendor/,,$@)

release: $(VENDOR) $(DIR_BIN)
	cd $(DIR_BIN); $(GO) build $(GO_FLAGS) $(DIR_PRE)

check: $(VENDOR)
	$(GO) build $(PKG)

test: check fmt lint vet
	$(GO) test $(GOTEST_FLAGS) $(PKG)

fmt:
	$(GOFMT) -l $(filter-out .,$(DIR)) $(wildcard *.go)

lint:
	$(GOLINT) -set_exit_status $(PKG)

vet:
	$(GO) vet $(PKG)

list:
	@echo $(PKG) | tr ' ' '\n'

clean:
	-rm -r $(DIR_BIN)
	go clean $(PKG)
	$(MAKE) -C $(GOW3)/vendor clean
