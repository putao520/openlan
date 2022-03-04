SHELL := /bin/bash

.ONESHELL:
.PHONY: linux linux-rpm darwin darwin-zip windows windows-zip test vendor

## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell cat VERSION)

## declare flags
MOD = github.com/danieldin95/openlan/pkg/libol
LDFLAGS += -X $(MOD).Commit=$(shell git rev-list -1 HEAD)
LDFLAGS += -X $(MOD).Date=$(shell date +%FT%T%z)
LDFLAGS += -X $(MOD).Version=$(VER)

## declare directory
SD = $(shell pwd)
BD = $(SD)/build
LD = openlan-linux-$(VER)
WD = openlan-windows-$(VER)
XD = openlan-darwin-$(VER)
DEST = $(DST)

build: test pkg

pkg: linux-rpm linux-tar windows-zip darwin-zip ## build all plaftorm packages

help: ## show make targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);\
	printf " \033[36m%-20s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all platform
bin: linux windows darwin ## build all platform binary

#
## docker run --network host --privileged -v /var/run:/var/run -v /etc/openlan/switch:/etc/openlan/switch openlan-switch:5.8.13
docker: pkg
	docker build -t openlan-switch:$(VER) --build-arg VERSION=$(VER) -f ./dist/openlan-switch.docker  .

## upgrade
upgrade:
	ansible-playbook ./misc/playbook/upgrade.yaml -e "version=$(VER)"

install:
	echo $(DEST)

clean: ## clean cache
	rm -rvf ./build
	rm -rvf ./core/build
	rm -rvf ./core/cmake-build-debug

## prepare environment
vendor:
	go clean -modcache
	go mod tidy
	go mod vendor -v

env:
	@mkdir -p $(BD)
	@go version
	@gofmt -w -s ./pkg ./cmd ./misc
	@git submodule init
	@git submodule update

## linux platform
linux: linux-proxy linux-point linux-switch

openudp: env
	mkdir -p ./build/udp
	cd ./build/udp && cmake $(SD)/core/udp && make

opentcp: env
	mkdir -p ./build/tcp
	cd ./build/tcp && cmake $(SD)/core/tcp && make

## compile command line
cmd: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan ./cmd/main.go
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan.i386 ./cmd/main.go

linux-bin: linux-point linux-switch linux-proxy ## build linux binary

linux-point: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point ./cmd/point_linux
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.i386 ./cmd/point_linux

linux-switch: env openudp
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch ./cmd/switch
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-switch.i386 ./cmd/switch

	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan ./cmd/main.go
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan.i386 ./cmd/main.go

linux-proxy: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy ./cmd/proxy
	GOARCH=386 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-proxy.i386 ./cmd/proxy

linux-rpm: env ## build rpm packages
	@dist/spec.sh
	rpmbuild -ba $(BD)/openlan.spec

linux-tar: env linux-point linux-switch linux-proxy ## build linux packages
	@pushd $(BD)
	@rm -rf $(LD) && mkdir -p $(LD)
	@rm -rf $(LD).tar

	@mkdir -p $(LD)/etc/sysctl.d
	@cp -rvf $(SD)/dist/resource/90-openlan.conf $(LD)/etc/sysctl.d
	@mkdir -p $(LD)/etc/openlan
	@cp -rvf $(SD)/dist/resource/point.json.example $(LD)/etc/openlan
	@cp -rvf $(SD)/dist/resource/proxy.json.example $(LD)/etc/openlan
	@mkdir -p $(LD)/etc/openlan/switch
	@cp -rvf $(SD)/dist/resource/switch.json.example $(LD)/etc/openlan/switch
	@mkdir -p $(LD)/etc/openlan/switch/network
	@cp -rvf $(SD)/dist/resource/network.json.example $(LD)/etc/openlan/switch/network
	@mkdir -p $(LD)/usr/bin
	@cp -rvf $(BD)/openudp $(LD)/usr/bin
	@cp -rvf $(BD)/openlan $(LD)/usr/bin
	@cp -rvf $(BD)/openlan-proxy $(LD)/usr/bin
	@cp -rvf $(BD)/openlan-point $(LD)/usr/bin
	@cp -rvf $(BD)/openlan-switch $(LD)/usr/bin
	@mkdir -p $(LD)/var/openlan
	@mkdir -p $(LD)/var/openlan/point
	@mkdir -p $(LD)/var/openlan/openvpn
	@cp -rvf $(SD)/dist/resource/cert/openlan/cert $(LD)/var/openlan
	@cp -rvf $(SD)/dist/script $(LD)/var/openlan
	@cp -rvf $(SD)/dist/resource/cert/openlan/ca/ca.crt $(LD)/var/openlan/cert
	@mkdir -p $(LD)/etc/sysconfig/openlan
	@cp -rvf $(SD)/dist/resource/point.cfg $(LD)/etc/sysconfig/openlan
	@cp -rvf $(SD)/dist/resource/proxy.cfg $(LD)/etc/sysconfig/openlan
	@cp -rvf $(SD)/dist/resource/switch.cfg $(LD)/etc/sysconfig/openlan
	@mkdir -p $(LD)//usr/lib/systemd/system
	@cp -rvf $(SD)/dist/resource/openlan-point@.service $(LD)/usr/lib/systemd/system
	@cp -rvf $(SD)/dist/resource/openlan-proxy.service $(LD)/usr/lib/systemd/system
	@cp -rvf $(SD)/dist/resource/openlan-switch.service $(LD)/usr/lib/systemd/system

	tar -cf $(LD).tar $(LD)
	@rm -rf $(LD)

## cross build for windows
windows: windows-point ## build windows binary

windows-point: env
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.exe ./cmd/point_windows

windows-zip: env windows ## build windows packages
	@pushd $(BD)
	@rm -rf $(WD) && mkdir -p $(WD)
	@rm -rf $(WD).zip

	@cp -rvf $(SD)/dist/resource/point.json.example $(WD)/point.json
	@cp -rvf $(BD)/openlan-point.exe $(WD)

	zip -r $(WD).zip $(WD) > /dev/null
	@rm -rf $(WD)

windows-syso: ## build windows syso
	rsrc -manifest ./cmd/point_windows/main.manifest -ico ./cmd/point_windows/main.ico  -o ./cmd/point_windows/main.syso

## cross build for osx
osx: darwin

darwin: env ## build darwin binary
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.darwin ./cmd/point_darwin

darwin-zip: env darwin ## build darwin packages
	@pushd $(BD)
	@rm -rf $(XD) && mkdir -p $(XD)
	@rm -rf $(XD).zip

	@cp -rvf $(SD)/dist/resource/point.json.example $(XD)/point.json
	@cp -rvf $(BD)/openlan-point.darwin $(XD)

	zip -r $(XD).zip $(XD) > /dev/null
	@rm -rf $(XD)

## unit test
test: ## execute unit test
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan/pkg/olap
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan/pkg/libol
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan/pkg/models
