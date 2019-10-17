prefix = /usr
bindir := $(prefix)/bin

VERSION := $(git describe --tags)
ARCHS := amd64
arch = $(word 1, $@)

# If the first argument is "run"...
ifeq (image,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  IMAGE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(IMAGE_ARGS):;@:)
endif

build:
	mkdir -p bin
	go build -o bin/gourmet cmd/main.go

.PHONY: image
image:
	docker build -t $(IMAGE_ARGS) .

install:
	install -p -m 0755 gourmet $(bindir)
	setcap cap_net_raw,cap_net_admin=eip $(bindir)/gourmet

.PHONY: $(ARCHS)
$(ARCHS):
	mkdir -p release
	GOARCH=$(arch) GOOS=linux go build -o gourmet cmd/main.go
	zip release/gourmet-$(VERSION)-$(arch).zip gourmet
	rm gourmet

release: $(ARCHS)

