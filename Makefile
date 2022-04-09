TARGET		:= k3se
SOURCES		:= $(shell find . -name "*.go")
PLATFORM	?= $(shell go version | cut -d " " -f 4)
GOOS		:= $(shell echo $(PLATFORM) | cut -d "/" -f 1)
GOARCH		:= $(shell echo $(PLATFORM) | cut -d "/" -f 2)
SUFFIX		:= $(GOOS)-$(GOARCH)
VERSION		?= $(shell git describe --always --tags --dirty)
BUILD_FLAGS	:= -ldflags="-s -w -X github.com/nicklasfrahm/$(TARGET)/cmd.version=$(VERSION)"

# Adjust the binary name on Windows.
ifeq ($(GOOS),windows)
SUFFIX	= $(GOOS)-$(GOARCH).exe
endif

build: bin/$(TARGET)-$(SUFFIX)

bin/$(TARGET)-$(SUFFIX): $(SOURCES)
	@mkdir -p $(@D)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $@ main.go
ifdef UPX
	upx -qq $(UPX) $@-$(SUFFIX)
endif

.PHONY: clean
clean:
	@rm -rvf bin
