# Author:  Niels A.D.
# Project: goop (https://github.com/nielsAD/goop)
# License: Mozilla Public License, v2.0

GOW3=third_party/gowarcraft3
THIRD_PARTY=$(GOW3)/third_party/bncsutil/build/libbncsutil.a

GO_LDFLAGS=
GO_FLAGS=
GOTEST_FLAGS=-cover -cpu=1,2,4 -timeout=2m

GO=go
GOFMT=gofmt
STATICCHECK=$(shell $(GO) env GOPATH)/bin/staticcheck

DIR_BIN=bin
DIR_PRE=github.com/nielsAD/goop

PKG:=$(shell $(GO) list ./...)
DIR:=$(subst $(DIR_PRE),.,$(PKG))

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

.PHONY: all release check test fmt lint vet list clean install-tools res.syso

all: test release

$(DIR_BIN):
	mkdir -p $@

$(PKG): $(THIRD_PARTY)
	$(GO) build $@

$(GOW3)/third_party/%:
	$(MAKE) -C $(GOW3)/third_party $(subst $(GOW3)/third_party/,,$@)

res.syso:
	echo "$$RES" | $(WINDRES) -c 65001 -O coff -o $@

release: $(THIRD_PARTY) $(DIR_BIN) res.syso
	cd $(DIR_BIN); $(GO) build $(GO_FLAGS) -ldflags '-X main.BuildTag=$(GIT_TAG) -X main.BuildCommit=$(GIT_COMMIT) -X main.buildDate=$(shell date +'%s') $(GO_LDFLAGS)' $(DIR_PRE)

check: $(THIRD_PARTY)
	$(GO) build $(PKG)

test: check fmt lint vet
	$(GO) test $(GOTEST_FLAGS) $(PKG)

fmt:
	@GOFMT_OUT=$$($(GOFMT) -d $(filter-out .,$(DIR)) $(wildcard *.go) 2>&1); \
	if [ -n "$$GOFMT_OUT" ]; then \
		echo "$$GOFMT_OUT"; \
		exit 1; \
	fi

lint:
	$(STATICCHECK) $(PKG)
	cd plugins; $(LUACHECK) .

vet:
	$(GO) vet $(PKG)

list:
	@echo $(PKG) | tr ' ' '\n'

clean:
	-rm -r $(DIR_BIN) res.syso
	$(GO) clean $(PKG)
	$(MAKE) -C $(GOW3)/third_party clean

install-tools:
	$(GO) mod download
	grep -o '"[^"]\+"' tools.go | xargs -n1 $(GO) install
