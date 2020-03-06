.PHONY: linux rpm win-zip test

PKG = github.com/danieldin95/openlan-go/config

LDFLAGS += -X $(PKG).Commit=$$(git rev-list -1 HEAD)
LDFLAGS += -X $(PKG).Date=$$(date +%FT%T%z)
LDFLAGS += -X $(PKG).Version=$$(cat VERSION)

WIN_DIR = "openlan-windows-"$$(cat VERSION)


linux: linux/point linux/vswitch


linux/point:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o point.linux.x86_64 main/point_linux.go


linux/vswitch:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o vswitch.linux.x86_64 main/vswitch.go


linux/rpm:
	./packaging/auto.sh
	rpmbuild -ba packaging/openlan-point.spec
	rpmbuild -ba packaging/openlan-vswitch.spec
	cp -rvf ~/rpmbuild/RPMS/x86_64/openlan-*.rpm .


windows:
	go build -mod=vendor -o point.windows.x86_64.exe main/point_windows.go
	go build -mod=vendor -o vswitch.windows.x86_64.exe main/vswitch.go



windows/zip:
	rm -rf $(WIN_DIR) && mkdir -p $(WIN_DIR)
	cp -rvf resource/point.json $(WIN_DIR)
	cp -rvf point.windows.x86_64.exe $(WIN_DIR)
	cp -rvf vswitch.windows.x86_64.exe $(WIN_DIR)
	rm -rf $(WIN_DIR).zip
	zip -r $(WIN_DIR).zip $(WIN_DIR)


osx: darwin



darwin:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o point.darwin.x86_64 main/point_darwin.go
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o vswitch.darwin.x86_64 main/vswitch.go


docker:
	docker build -t openlan-point -f packaging/point/Dockerfile .
	docker build -t openlan-vswitch -f packaging/vswitch/Dockerfile .
	# --env VS_ADDR=192.168.209.141 --env VS_AUTH=hi@admin:hi123$ --env VS_TLS=true
	# docker run -d --privileged openlan-point:latest
	# docker run -d  -p 10000:10000 -p 10002:10002 openlan-vswitch:latest


test:
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/point
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/libol
