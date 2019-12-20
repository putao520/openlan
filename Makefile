.PHONY: linux rpm win-zip

linux:
	go build -o ./resource/point.linux.x86_64 main/point_linux.go
	go build -o ./resource/vswitch.linux.x86_64 main/vswitch_linux.go

windows:
	go build -o ./resource/point.windows.x86_64 main/point_windows.go

osx: darwin

darwin:
	go build -o ./resource/point.darwin.x86_64 main/point_darwin.go

rpm:
	./packaging/auto.sh
	rpmbuild -ba ./packaging/openlan-point.spec
	rpmbuild -ba ./packaging/openlan-vswitch.spec
	cp -rvf ~/rpmbuild/RPMS/x86_64/openlan-* ./resource

win-zip:
	mkdir -p ./openlan-wins
	cp -rvf ./resource/point.windows.x86_64 ./openlan-wins/point.windows.x86_64.exe
	cp -rvf ./resource/point.json ./openlan-wins/point.json
	zip -r ./resource/openlan-wins.zip ./openlan-wins
