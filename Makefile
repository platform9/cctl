# Copyright (c) 2018 Platform9 Systems, Inc.
#
# Usage:
# make                 # builds the artifact
# make clean           # removes the artifact and the vendored packages
# make clean-all       # same as make clean + removes the bin dir which houses dep
# make container-build # build artifact on a Linux based container using the latest golang

SHELL := /usr/bin/env bash
BUILD_NUMBER ?= 10
GITHASH := $(shell git rev-parse --short HEAD)
CWD := $(shell pwd)
PF9_VERSION ?= 5.5.0
SRC_DIR := $(CWD)/src/pf9-clusteradm
VERSION := $(PF9_VERSION)-$(BUILD_NUMBER)
DETECTED_OS := $(shell uname -s)
DEP_BIN_GIT := https://github.com/golang/dep/releases/download/v0.4.1/dep-$(DETECTED_OS)-amd64
DEP_BIN := $(CWD)/bin/dep
BIN := pf9-clusteradm

.PHONY: clean clean-all gopath depnolock container-build

export GOPATH=$(CWD)
export DEPNOLOCK=1 # issue with vboxsf (vagrant + vbox) : https://github.com/golang/dep/issues/947

default: $(BIN)

container-build:
	docker run --rm -v $(PWD):/build -w /build golang:latest make

$(DEP_BIN):
	mkdir $(CWD)/bin
	wget $(DEP_BIN_GIT) -O $@
	chmod +x $(DEP_BIN)

$(BIN):  $(DEP_BIN)
	pushd $(SRC_DIR) &&\
	$(DEP_BIN) ensure &&\
	go build main.go &&\
	mv main $(BIN)

clean-all: clean
	rm -rf bin
clean:
	rm -rf build pkg $(SRC_DIR)/{vendor,${BIN}}
