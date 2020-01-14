.PHONY: linux rpm win-zip test

linux:
	go build -mod=vendor -o ./resource/point.linux.x86_64 main/point_linux.go
	go build -mod=vendor -o ./resource/vswitch.linux.x86_64 main/vswitch.go

windows:
	go build -mod=vendor -o ./resource/point.windows.x86_64.exe main/point_windows.go
	go build -mod=vendor -o ./resource/vswitch.windows.x86_64.exe main/vswitch.go

osx: darwin

darwin:
	go build -mod=vendor -o ./resource/point.darwin.x86_64 main/point_darwin.go
	go build -mod=vendor -o ./resource/vswitch.darwin.x86_64 main/vswitch.go

rpm:
	./packaging/auto.sh
	rpmbuild -ba ./packaging/openlan-point.spec
	rpmbuild -ba ./packaging/openlan-vswitch.spec
	cp -rvf ~/rpmbuild/RPMS/x86_64/openlan-* ./resource

win-zip:
	mkdir -p ./openlan-wins
	cp -rvf ./resource/point.windows.x86_64 ./openlan-wins
	cp -rvf ./resource/point.json ./openlan-wins
	zip -r ./resource/openlan-wins.zip ./openlan-wins

test:
	go test -mod=vendor github.com/danieldin95/openlan-go/point
	go test -mod=vendor -bench=. github.com/danieldin95/openlan-go/point
