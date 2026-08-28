APP_NAME := taxiprijs
VERSION  := 1.0.1
GOFLAGS  := -trimpath
LDFLAGS  := -s -w -X main.version=$(VERSION)

# Cross-compile targets
.PHONY: build build-linux build-macos build-windows install uninstall clean

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP_NAME) ./cmd/taxiprijs

build-linux:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux-amd64 ./cmd/taxiprijs
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux-arm64 ./cmd/taxiprijs

build-macos:
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-macos-amd64 ./cmd/taxiprijs
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-macos-arm64 ./cmd/taxiprijs

build-windows:
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-windows-amd64.exe ./cmd/taxiprijs

build-all: build-linux build-macos build-windows
	@echo "Builds complete in dist/"

install: build
	sudo install -Dm755 $(APP_NAME) /usr/local/bin/$(APP_NAME)
	sudo install -Dm644 extras/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop
	@mkdir -p $$HOME/.config/omarchy/plugins/jp.taxiprijs
	@cp extras/omarchy-plugin/* $$HOME/.config/omarchy/plugins/jp.taxiprijs/
	@if [ -f "$$HOME/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq --arg id jp.taxiprijs 'if ([.bar.layout.center[]? , .bar.layout.left[]? , .bar.layout.right[]?] | any(.id == $$id)) then . else .bar.layout.right = ((.bar.layout.right // []) + [{"id": $$id}]) end' "$$HOME/.config/omarchy/shell.json" > "$$HOME/.config/omarchy/shell.json.tmp" && mv "$$HOME/.config/omarchy/shell.json.tmp" "$$HOME/.config/omarchy/shell.json"; \
	fi
	@if command -v omarchy-shell >/dev/null 2>&1; then omarchy-shell -q shell rescanPlugins || true; fi
	@echo "Installed: /usr/local/bin/$(APP_NAME)"
	@echo "Desktop entry: /usr/share/applications/$(APP_NAME).desktop"
	@echo "Omarchy QML plugin: $$HOME/.config/omarchy/plugins/jp.taxiprijs"

uninstall:
	sudo rm -f /usr/local/bin/$(APP_NAME)
	sudo rm -f /usr/share/applications/$(APP_NAME).desktop
	sudo rm -f /usr/local/share/man/man1/$(APP_NAME).1
	rm -rf $$HOME/.$(APP_NAME)
	rm -rf $$HOME/.config/omarchy/plugins/jp.taxiprijs
	@if [ -f "$$HOME/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq '(.bar.layout.center // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.left // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.right // []) |= map(select(.id != "jp.taxiprijs"))' "$$HOME/.config/omarchy/shell.json" > "$$HOME/.config/omarchy/shell.json.tmp" && mv "$$HOME/.config/omarchy/shell.json.tmp" "$$HOME/.config/omarchy/shell.json"; \
	fi
	@echo "Uninstalled"

clean:
	rm -rf dist/ $(APP_NAME)
