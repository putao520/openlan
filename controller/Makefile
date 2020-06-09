#
# github.com/danieldin95/openlan-go/controller
#

SHELL := /bin/bash

.ONESHELL:
.PHONY: linux test


## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell cat ../VERSION)

## declare flags
MOD = github.com/danieldin95/openlan-go/libol
LDFLAGS += -X $(MOD).Commit=$(shell git rev-list -1 HEAD)
LDFLAGS += -X $(MOD).Date=$(shell date +%FT%T%z)
LDFLAGS += -X $(MOD).Version=$(VER)

## declare directory
SD = $(shell pwd)
BD = $(SD)/../build

all: linux

env:
	@mkdir -p $(BD)

linux:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-ctrl main.go
