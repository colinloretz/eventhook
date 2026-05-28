VERSION ?= 0.1.0
LDFLAGS  = -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: build dashboard dev test release

build: dashboard
	go build $(LDFLAGS) -o bin/eventhook ./cmd/eventhook

dashboard:
	cd dashboard && npm run build
	rm -rf assets/dashboard
	mkdir -p assets/dashboard
	cp -r dashboard/dist/* assets/dashboard/

dev: build
	./bin/eventhook dev

test:
	go test ./...

# Cross-compile for all platforms (used for GitHub releases)
release: dashboard
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/eventhook_darwin_arm64  ./cmd/eventhook
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/eventhook_darwin_amd64  ./cmd/eventhook
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o dist/eventhook_linux_arm64   ./cmd/eventhook
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/eventhook_linux_amd64   ./cmd/eventhook
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o dist/eventhook_windows_amd64.exe ./cmd/eventhook
	cd dist && for f in eventhook_darwin_* eventhook_linux_*; do tar czf $$f.tar.gz $$f && rm $$f; done
	@echo "Release artifacts in dist/"
