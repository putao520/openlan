.help:
	@echo "make win      building point on windows"
	@echo "make linux    building point and vswitch on linux"
	@echo "make install  install openlan to linux"

linux:
	go build -o ./resource/point point_linux.go
	go build -o ./resource/vswitch vswitch_linux.go

install:
	./install.sh

win:
	go build -o ./resource/point.exe point_windows.go
