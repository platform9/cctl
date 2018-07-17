# Copyright (c) 2018 Platform9 Systems, Inc.
#
# Usage:
# make                 # builds the artifact
# make ensure          # runs dep ensure which downloads the dependencies
# make clean           # removes the artifact and the vendored packages
# make clean-all       # same as make clean + removes the bin dir which houses dep
# make container-build # build artifact on a Linux based container using golang 1.10

SHELL := /usr/bin/env bash
BUILD_NUMBER ?= 10
GITHASH := $(shell git rev-parse --short HEAD)
CWD := $(shell pwd)
PF9_VERSION ?= 5.5.0
VERSION := $(PF9_VERSION)-$(BUILD_NUMBER)
DETECTED_OS := $(shell uname -s)
DEP_BIN_GIT := https://github.com/golang/dep/releases/download/v0.4.1/dep-$(DETECTED_OS)-amd64
BIN := cctl
REPO := cctl
PACKAGE_GOPATH := /go/src/github.com/platform9/$(REPO)
DEP_TEST=$(shell which dep)

ifeq ($(DEP_TEST),)
	DEP_BIN := $(CWD)/bin/dep
else
	DEP_BIN := $(DEP_TEST)
endif

.PHONY: clean clean-all container-build default ensure format

default: $(BIN)

container-build:
	docker run --rm -v $(PWD):$(PACKAGE_GOPATH) -w $(PACKAGE_GOPATH) golang:1.10 make

$(DEP_BIN):
ifeq ($(DEP_BIN),$(CWD)/bin/dep)
	echo "Downloading dep from GitHub" &&\
	mkdir -p $(CWD)/bin &&\
	wget $(DEP_BIN_GIT) -O $(DEP_BIN) &&\
	chmod +x $(DEP_BIN)
endif

ensure: $(DEP_BIN)
	echo $(DEP_BIN)
	$(DEP_BIN) ensure -v

$(BIN):
	go build -o $(BIN)

format:
	gofmt -w -s *.go
	gofmt -w -s */*.go

clean-all: clean
	rm -rf bin

clean:
	rm -rf $(BIN)
