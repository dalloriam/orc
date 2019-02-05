# Makefile magic from Jessie Frazelle:
# https://github.com/jessfraz

NAME := orc
PREFIX ?=$(shell pwd)
BUILDDIR := ${PREFIX}/dist

VERSION := $(shell cat VERSION.txt)

GOOSARCHES = linux/amd64 darwin/amd64 windows/amd64
BUILDTAGS :=


CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

GO := go

define buildpretty
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS) static_build netgo" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} ./client;
md5sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).md5;
sha256sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
zip $(BUILDDIR)/$(NAME)-$(1)-$(2).zip $(BUILDDIR)/$(NAME)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(1)-$(2).md5 $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
rm $(BUILDDIR)/$(NAME)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(1)-$(2).md5 $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: build
build: