.help:
	@echo "make darwin   building point on macOS"
	@echo "make windows  building point on windows"
	@echo "make linux    building point and vswitch on linux"
	@echo "make install  install openlan to linux"

linux:
	go build -o ./resource/point point_linux.go
	go build -o ./resource/vswitch vswitch_linux.go
	go build -o ./resource/pointctl point_ctl.go

windows:
	go build -o ./resource/point.exe point_windows.go

darwin:
	go build -o ./resource/point.dw point_darwin.go

install:
	./install.sh
