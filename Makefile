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
TMP_DIR := /tmp/pf9-clusteradm
TMP_SRC_DIR := $(TMP_DIR)/src
VERSION := $(PF9_VERSION)-$(BUILD_NUMBER)
DETECTED_OS := $(shell uname -s)
DEP_BIN_GIT := https://github.com/golang/dep/releases/download/v0.4.1/dep-$(DETECTED_OS)-amd64
DEP_BIN := $(CWD)/bin/dep
BIN := pf9-clusteradm

.PHONY: clean clean-all gopath depnolock container-build

export GOPATH=$(TMP_DIR)
export DEPNOLOCK=1 # issue with vboxsf (vagrant + vbox) : https://github.com/golang/dep/issues/947

default: $(BIN)

container-build:
	docker run --rm -v $(PWD):/build -w /build golang:latest make

$(DEP_BIN):
	mkdir $(CWD)/bin
	wget $(DEP_BIN_GIT) -O $@
	chmod +x $(DEP_BIN)

$(BIN):  $(DEP_BIN)
	mkdir -p $(TMP_SRC_DIR)
	if [ ! -L $(TMP_SRC_DIR)/pf9-clusteradm ]; then\
		ln -s $(CWD) $(TMP_SRC_DIR)/pf9-clusteradm;\
	fi
	pushd $(TMP_SRC_DIR)/pf9-clusteradm &&\
	$(DEP_BIN) ensure -v &&\
	go build main.go &&\
	mv main $(BIN)

clean-all: clean
	rm -rf bin $(TMP_DIR)
clean:
	rm -rf build
