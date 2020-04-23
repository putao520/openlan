#
# github.com/danieldin95/openlan-go/controller
#

.PHONY: linux test

PKG = github.com/danieldin95/openlan-go/libol

LDFLAGS += -X $(PKG).Commit=$$(git rev-list -1 HEAD)
LDFLAGS += -X $(PKG).Date=$$(date +%FT%T%z)
LDFLAGS += -X $(PKG).Version=$$(cat ../VERSION)

all: linux

linux:
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o openlan-ctrl main.go
