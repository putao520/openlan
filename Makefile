#
# github.com/danieldin95/openlan-go
#

SHELL := /bin/bash

.ONESHELL:
.PHONY: linux linux-rpm darwin darwin-zip windows windows-zip test

## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell cat VERSION)

## declare flags
MOD = github.com/danieldin95/openlan-go/src/libol
LDFLAGS += -X $(MOD).Commit=$(shell git rev-list -1 HEAD)
LDFLAGS += -X $(MOD).Date=$(shell date +%FT%T%z)
LDFLAGS += -X $(MOD).Version=$(VER)

## declare directory
SD = $(shell pwd)
BD = $(SD)/build
LD = openlan-$(LSB)-$(VER)
WD = openlan-Windows-$(VER)
XD = openlan-Darwin-$(VER)

help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);\
		printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all platform
bin: linux windows darwin ## build all platform binary

pkg: linux-rpm windows-zip darwin-zip ## build all plaftorm packages

## upgrade
upgrade:
	ansible-playbook ./misc/playbook/upgrade.yaml -e "version=$(VER)"

clean: ## clean cache
	rm -rvf ./build
	rm -rvf ./core/build
	rm -rvf ./core/cmake-build-debug

## prepare environment
env:
	@mkdir -p $(BD)
	gofmt -w -s ./src

## linux platform
linux: linux-point linux-switch linux-ctrl ## build linux binary

linux-ctrl: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-ctrl ./src/cli/ctrl
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-ctrl ./src/cli/ctrl

linux-point: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point ./src/cli/point_linux
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.i386 ./src/cli/point_linux

linux-switch: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch ./src/cli/switch
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch.i386 ./src/cli/switch

linux-rpm: env ## build rpm packages
	@./packaging/spec.sh
	rpmbuild -ba packaging/openlan-ctrl.spec
	rpmbuild -ba packaging/openlan-point.spec
	rpmbuild -ba packaging/openlan-switch.spec
	@cp -rf ~/rpmbuild/RPMS/x86_64/openlan-*.rpm $(BD)

## cross build for windows
windows: windows-point ## build windows binary

windows-point: env
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.exe ./src/cli/point_windows

windows-zip: env windows ## build windows packages
	@pushd $(BD)
	@rm -rf $(WD) && mkdir -p $(WD)
	@rm -rf $(WD).zip

	@cp -rvf $(SD)/packaging/resource/point.json.example $(WD)/point.json
	@cp -rvf $(BD)/openlan-point.exe $(WD)

	zip -r $(WD).zip $(WD) > /dev/null
	@rm -rf $(WD)
	@popd

windows-syso: ## build windows syso
	rsrc -manifest main/point_windows-main.manifest -ico ./src/cli/point_windows-main.ico  -o ./src/cli/point_windows-main.syso

## cross build for osx
osx: darwin

darwin: env ## build darwin binary
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.darwin ./src/cli/point_darwin

darwin-zip: env darwin ## build darwin packages
	@pushd $(BD)
	@rm -rf $(XD) && mkdir -p $(XD)
	@rm -rf $(XD).zip

	@cp -rvf $(SD)/packaging/resource/point.json.example $(XD)/point.json
	@cp -rvf $(BD)/openlan-point.darwin $(XD)

	zip -r $(XD).zip $(XD) > /dev/null
	@rm -rf $(XD)
	popd

## unit test
test: ## execute unit test
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/src/point
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/src/libol
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/src/models
