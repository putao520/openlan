.help:
	@echo "make darwin   building point on macOS"
	@echo "make windows  building point on windows"
	@echo "make linux    building point and vswitch on linux"
	@echo "make install  install openlan to linux"

linux:
	go build -o ./resource/point.linux.x86_64 point_linux.go
	go build -o ./resource/vswitch.linux.x86_64 vswitch_linux.go
	go build -o ./resource/pointctl.linux.x86_64 pointctl.go

windows:
	go build -o ./resource/point.windows.x86_64 point_windows.go

darwin:
	go build -o ./resource/point.darwin.x86_64 point_darwin.go
	go build -o ./resource/pointctl.darwin.x86_64 pointctl.go

install:
	./install.sh
