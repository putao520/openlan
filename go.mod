module github.com/danieldin95/openlan-go

go 1.12

require (
	github.com/akavel/rsrc v0.8.0 // indirect
	github.com/danieldin95/lightstar v0.0.0-20200401145448-034e11afcf81
	github.com/gorilla/mux v1.7.4
	github.com/songgao/water v0.0.0-20190725173103-fd331bda3f4b
	github.com/stretchr/testify v1.5.1
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	golang.org/x/sys v0.0.0 // indirect
)

replace golang.org/x/net v0.0.0 => github.com/golang/net v0.0.0-20190812203447-cdfb69ac37fc

replace golang.org/x/sys v0.0.0 => github.com/golang/sys v0.0.0-20190209173611-3b5209105503
