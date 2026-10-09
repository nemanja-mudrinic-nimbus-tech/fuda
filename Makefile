SHELL := /bin/bash
ENV := set -a; [ -f .env ] && . ./.env; set +a;

.PHONY: dev dev-api dev-client check check-api check-client fmt build client docker desktop desktop-macos desktop-windows clean

VERSION ?= dev
GITHUB_CLIENT_ID ?=
AZURE_CLIENT_ID ?=
ARCH ?= $(shell go env GOARCH)
DESKTOP_IDS := $(ENV) github_id="$(GITHUB_CLIENT_ID)"; azure_id="$(AZURE_CLIENT_ID)";
DESKTOP_LDFLAGS := -X main.version=$(VERSION) -X main.githubClientID=$${github_id:-$$FUDA_GITHUB_CLIENT_ID} -X main.azureClientID=$${azure_id:-$$FUDA_AZURE_CLIENT_ID}

dev:
	$(MAKE) -j2 dev-api dev-client

dev-api:
	go -C api build -o ../bin/fuda ./cmd/fuda
	$(ENV) ./bin/fuda

dev-client: client/node_modules
	pnpm -C client dev

check: check-api check-client

check-api:
	cd api && golangci-lint run ./... && go test ./...

check-client: client/node_modules
	pnpm -C client lint
	pnpm -C client fmt:check
	pnpm -C client typecheck
	pnpm -C client test

fmt: client/node_modules
	cd api && golangci-lint fmt ./...
	pnpm -C client fmt

client: client/node_modules
	pnpm -C client build

build: client
	go -C api build -o ../bin/fuda ./cmd/fuda

desktop: desktop-macos desktop-windows

desktop-macos: client
	$(DESKTOP_IDS) \
	CGO_ENABLED=1 GOOS=darwin GOARCH=$(ARCH) MACOSX_DEPLOYMENT_TARGET=13.0 CGO_LDFLAGS=-mmacosx-version-min=13.0 \
		go -C api build -trimpath -ldflags="-s -w $(DESKTOP_LDFLAGS)" -o ../bin/fuda.app/Contents/MacOS/fuda-desktop ./cmd/fuda-desktop
	mkdir -p bin/fuda.app/Contents/Resources
	cp api/cmd/fuda-desktop/icon.icns bin/fuda.app/Contents/Resources/icon.icns
	sed 's/@VERSION@/$(VERSION)/' api/cmd/fuda-desktop/Info.plist > bin/fuda.app/Contents/Info.plist

desktop-windows: client
	$(DESKTOP_IDS) \
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go -C api build -trimpath -ldflags="-s -w -H=windowsgui $(DESKTOP_LDFLAGS)" -o ../bin/fuda-desktop.exe ./cmd/fuda-desktop

docker:
	docker build -t fuda .

clean:
	rm -rf bin .fuda-cache
	find api/internal/web/dist -mindepth 1 ! -name .gitkeep -delete

client/node_modules: client/package.json client/pnpm-lock.yaml
	pnpm -C client install --frozen-lockfile
	touch client/node_modules
