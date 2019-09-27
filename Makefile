prefix = /usr
bindir := $(prefix)/bin

build:
	mkdir -p bin
	go build -o bin/gourmet cmd/main.go

install:
	install -p -m 0755 gourmet $(bindir)
	setcap cap_net_raw,cap_net_admin=eip $(bindir)/gourmet

VERSION := 0.1.0
ARCHS := amd64 386
arch = $(word 1, $@)

.PHONY: $(ARCHS)
$(ARCHS):
	mkdir -p release
	GOARCH=$(arch) GOOS=linux go build -o gourmet cmd/main.go
	zip release/gourmet-$(VERSION)-$(arch).zip gourmet
	rm gourmet

release: $(ARCHS)