# Author:  Niels A.D.
# Project: goop (https://github.com/nielsAD/goop)
# License: Mozilla Public License, v2.0

GOW3=vendor/github.com/nielsAD/gowarcraft3
VENDOR=$(GOW3)/vendor/bncsutil/build/libbncsutil_static.a

GO_LDFLAGS=
GO_FLAGS=
GOTEST_FLAGS=-cover -cpu=1,2,4 -timeout=2m

GO=go
GOFMT=gofmt
GOLINT:=$(shell $(GO) env GOPATH)/bin/golint

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

LUACHECK:=$(shell command -v luacheck 2>/dev/null)
ifndef LUACHECK
	LUACHECK=: LUACHECK NOT INSTALLED, SKIPPING luacheck
endif

WINDRES:=$(shell command -v windres 2>/dev/null)
ifndef WINDRES
	WINDRES=: WINDRES NOT INSTALLED, SKIPPING windres
endif

GIT=git
GIT_TAG:=$(shell $(GIT) describe --abbrev=0 --tags)
GIT_COMMIT:=$(shell $(GIT) rev-parse HEAD)

comma=,
VERSION:=$(subst v,,$(subst .,$(comma),$(GIT_TAG)),$(shell date +'%y%m'))

define RES
1 VERSIONINFO
FILEVERSION     $(VERSION)
PRODUCTVERSION  $(VERSION)
FILEFLAGSMASK   0X3FL
FILEFLAGS       0L
FILEOS          0X40004L
FILETYPE        0X1
FILESUBTYPE     0
BEGIN
	BLOCK "StringFileInfo"
	BEGIN
		BLOCK "040904B0"
		BEGIN
			VALUE "CompanyName", "nielsAD"
			VALUE "FileDescription", "Goop is a BNCS Channel Operator."
			VALUE "FileVersion", "$(GIT_TAG)"
			VALUE "InternalName", "goop"
			VALUE "LegalCopyright", "© nielsAD. All rights reserved."
			VALUE "OriginalFilename", "goop.exe"
			VALUE "ProductName", "Goop"
			VALUE "ProductVersion", "$(GIT_TAG)"
		END
	END
	BLOCK "VarFileInfo"
	BEGIN
			VALUE "Translation", 0x0409, 0x04B0
	END
END
1 ICON "goop.ico"
endef
export RES

.PHONY: all release check test fmt lint vet list clean res.syso

all: test release

$(DIR_BIN):
	mkdir -p $@

$(PKG): $(VENDOR)
	$(GO) build $@

$(GOW3)/vendor/%:
	$(MAKE) -C $(GOW3)/vendor $(subst $(GOW3)/vendor/,,$@)

res.syso:
	echo "$$RES" | $(WINDRES) -c 65001 -O coff -o $@

release: $(VENDOR) $(DIR_BIN) res.syso
	cd $(DIR_BIN); $(GO) build $(GO_FLAGS) -ldflags '-X main.BuildTag=$(GIT_TAG) -X main.BuildCommit=$(GIT_COMMIT) -X main.buildDate=$(shell date +'%s') $(GO_LDFLAGS)' $(DIR_PRE)

check: $(VENDOR)
	$(GO) build $(PKG)

test: check fmt lint vet
	$(GO) test $(GOTEST_FLAGS) $(PKG)

fmt:
	$(GOFMT) -l $(filter-out .,$(DIR)) $(wildcard *.go)

lint:
	$(GOLINT) -set_exit_status $(PKG)
	cd plugins; $(LUACHECK) .

vet:
	$(GO) vet $(PKG)

list:
	@echo $(PKG) | tr ' ' '\n'

clean:
	-rm -r $(DIR_BIN) res.syso
	go clean $(PKG)
	$(MAKE) -C $(GOW3)/vendor clean
