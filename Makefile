APP_NAME := taxiprijs
VERSION  := 1.0.1
GOFLAGS  := -trimpath
LDFLAGS  := -s -w -X main.version=$(VERSION)

# Cross-compile targets
INSTALL_DIR      := /usr/share/applications
ICON_DIR         := /usr/share/icons/hicolor/scalable/apps

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
	sudo install -Dm644 extras/$(APP_NAME).desktop $(INSTALL_DIR)/$(APP_NAME).desktop
	sudo install -Dm644 extras/$(APP_NAME).svg $(ICON_DIR)/$(APP_NAME).svg
	@command -v gtk-update-icon-cache >/dev/null 2>&1 && sudo gtk-update-icon-cache -f /usr/share/icons/hicolor || true
	@command -v update-desktop-database >/dev/null 2>&1 && sudo update-desktop-database $(INSTALL_DIR) || true
	@echo "Installed: /usr/local/bin/$(APP_NAME)"
	@echo "Desktop entry: $(INSTALL_DIR)/$(APP_NAME).desktop"
	@echo "Icon: $(ICON_DIR)/$(APP_NAME).svg"

uninstall:
	sudo rm -f /usr/local/bin/$(APP_NAME)
	sudo rm -f $(INSTALL_DIR)/$(APP_NAME).desktop
	sudo rm -f $(ICON_DIR)/$(APP_NAME).svg
	sudo rm -f /usr/local/share/man/man1/$(APP_NAME).1
	rm -rf $$HOME/.$(APP_NAME)
	rm -rf $$HOME/.config/omarchy/plugins/jp.taxiprijs
	@if [ -f "$$HOME/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq '(.bar.layout.center // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.left // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.right // []) |= map(select(.id != "jp.taxiprijs"))' "$$HOME/.config/omarchy/shell.json" > "$$HOME/.config/omarchy/shell.json.tmp" && mv "$$HOME/.config/omarchy/shell.json.tmp" "$$HOME/.config/omarchy/shell.json"; \
	fi
	@command -v gtk-update-icon-cache >/dev/null 2>&1 && sudo gtk-update-icon-cache -f /usr/share/icons/hicolor || true
	@command -v update-desktop-database >/dev/null 2>&1 && sudo update-desktop-database $(INSTALL_DIR) || true
	@echo "Uninstalled"

clean:
	rm -rf dist/ $(APP_NAME)
