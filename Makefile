#
# github.com/danieldin95/openlan-go
#

.PHONY: linux linux/rpm darwin darwin/zip windows windows/zip test

## version
PKG = github.com/danieldin95/openlan-go/config

LDFLAGS += -X $(PKG).Commit=$$(git rev-list -1 HEAD)
LDFLAGS += -X $(PKG).Date=$$(date +%FT%T%z)
LDFLAGS += -X $(PKG).Version=$$(cat VERSION)

# all platform

all: linux windows darwin

all/pkg: linux/rpm windows/zip darwin/zip

## linux platform
linux: linux/point linux/vswitch

linux/point:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o point.linux.x86_64 ./main/point_linux

linux/vswitch:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o vswitch.linux.x86_64 ./main/vswitch.go

linux/rpm:
	@./packaging/auto.sh
	rpmbuild -ba packaging/openlan-point.spec
	rpmbuild -ba packaging/openlan-vswitch.spec
	@cp -rvf ~/rpmbuild/RPMS/x86_64/openlan-*.rpm .

## cross build for windows
WIN_DIR = "openlan-windows-"$$(cat VERSION)

windows: windows/point windows/vswitch

windows/point:
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o point.windows.x86_64.exe ./main/point_windows

windows/vswitch:
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o vswitch.windows.x86_64.exe ./main/vswitch.go

windows/zip: windows
	@rm -rf $(WIN_DIR) && mkdir -p $(WIN_DIR)
	@cp -rvf packaging/resource/point.json $(WIN_DIR)
	@cp -rvf point.windows.x86_64.exe $(WIN_DIR)
	@cp -rvf vswitch.windows.x86_64.exe $(WIN_DIR)
	@rm -rf $(WIN_DIR).zip
	zip -r $(WIN_DIR).zip $(WIN_DIR)
	@rm -rf $(WIN_DIR)

windows/syso:
	rsrc -manifest main/point_windows/main.manifest -ico main/point_windows/main.ico  -o main/point_windows/main.syso

## cross build for osx

DARWIN_DIR = "openlan-darwin-"$$(cat VERSION)

osx: darwin

darwin:
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o point.darwin.x86_64 ./main/point_darwin
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o vswitch.darwin.x86_64 ./main/vswitch.go

darwin/zip: darwin
	@rm -rf $(DARWIN_DIR) && mkdir -p $(DARWIN_DIR)
	@cp -rvf packaging/resource/point.json $(DARWIN_DIR)
	@cp -rvf point.darwin.x86_64 $(DARWIN_DIR)
	@cp -rvf vswitch.darwin.x86_64 $(DARWIN_DIR)
	rm -rf $(DARWIN_DIR).zip
	@zip -r $(DARWIN_DIR).zip $(DARWIN_DIR)
	@rm -rf $(DARWIN_DIR)

## docker images
docker:
	docker build -t openlan-point -f packaging/point/Dockerfile .
	docker build -t openlan-vswitch -f packaging/vswitch/Dockerfile .
	# --env VS_ADDR=192.168.209.141 --env VS_AUTH=hi@admin:hi123$ --env VS_TLS=true
	# docker run -d --privileged openlan-point:latest
	# docker run -d  -p 10000:10000 -p 10002:10002 openlan-vswitch:latest

## unit test
test:
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/point
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/libol
