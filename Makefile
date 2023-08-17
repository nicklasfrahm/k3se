TARGET		:= k3se
SOURCES		:= $(shell find . -name "*.go")
PLATFORM	?= $(shell go version | cut -d " " -f 4)
GOOS		:= $(shell echo $(PLATFORM) | cut -d "/" -f 1)
GOARCH		:= $(shell echo $(PLATFORM) | cut -d "/" -f 2)
SUFFIX		:= $(GOOS)-$(GOARCH)
VERSION		?= $(shell git describe --always --tags --dirty)
BUILD_FLAGS	:= -ldflags="-s -w -X github.com/nicklasfrahm/$(TARGET)/cmd.version=$(VERSION)"
DESTDIR		:= /usr/local/bin

# Adjust the binary name on Windows.
ifeq ($(GOOS),windows)
SUFFIX	= $(GOOS)-$(GOARCH).exe
endif

build: bin/$(TARGET)-$(SUFFIX)

bin/$(TARGET)-$(SUFFIX): $(SOURCES)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $@ main.go
ifdef UPX
	upx -qq $(UPX) $@
endif

.PHONY: vagrant-up
vagrant-up:
	cd deploy/vagrant; vagrant up

.PHONY: vagrant-down
vagrant-down:
	cd deploy/vagrant; vagrant destroy -f

$(DESTDIR)/$(TARGET): bin/$(TARGET)-$(SUFFIX)
	sudo install -Dm 755 $^ $@

.PHONY: install
install: $(DESTDIR)/$(TARGET)

.PHONY: uninstall
uninstall:
	@sudo rm -f $(DESTDIR)/$(TARGET)

.PHONY: docker
docker:
	docker build \
	  -t $(TARGET):latest \
	  -t $(TARGET):$(VERSION) \
	  --build-arg VERSION=$(VERSION) \
	  -f build/package/Dockerfile .

.PHONY: clean
clean:
	@rm -rvf bin

.PHONY: demo-up
demo-up: install
	@echo -n "\e[35m==>\e[0m "
	k3se up deploy/demo/k3se.yaml
	@echo -n "\e[35m==>\e[0m "
	kubectx admin@k3se.nicklasfrahm.xyz
	@echo -n "\e[35m==>\e[0m "
	kubectl create ns traefik --dry-run=client -o yaml | kubectl apply -f -
	@echo -n "\e[35m==>\e[0m "
	helm dependency update deploy/demo/traefik
	@echo -n "\e[35m==>\e[0m "
	helm upgrade --install traefik deploy/demo/traefik --namespace traefik
	@echo -n "\e[35m==>\e[0m "
	kubectl create ns cert-manager --dry-run=client -o yaml | kubectl apply -f -
	@echo -n "\e[35m==>\e[0m "
	helm dependency update deploy/demo/cert-manager
	@echo -n "\e[35m==>\e[0m "
	helm upgrade --install cert-manager deploy/demo/cert-manager --namespace cert-manager
	@echo -n "\e[35m==>\e[0m "
	kubectl apply -f deploy/demo/clusterissuers
	@echo -n "\e[35m==>\e[0m "
	kubectl apply -f deploy/demo/app

.PHONY: demo-down
demo-down:
	@echo -n "\e[35m==>\e[0m "
	k3se down deploy/demo/k3se.yaml
