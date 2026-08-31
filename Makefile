APP_NAME := taxiprijs
VERSION  := 1.1.0
GOFLAGS  := -trimpath
LDFLAGS  := -s -w -X main.version=$(VERSION)
INSTALLER_NAME := install-taxicheck-v$(VERSION)-windows.exe
PLUGIN_ID := jp.taxiprijs
PLUGIN_DST := $(HOME)/.config/omarchy/plugins/$(PLUGIN_ID)

# Cross-compile targets
.PHONY: build build-linux build-macos build-windows build-windows-installer install install-user install-system uninstall clean

# Build to a temp name then rename so this works while taxiprijs is running
# (in-place go build -o taxiprijs fails with ETXTBSY).
build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o .$(APP_NAME).new ./cmd/taxiprijs
	chmod +x .$(APP_NAME).new
	mv -f .$(APP_NAME).new $(APP_NAME)

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

# Self-contained Windows installer. Embeds taxiprijs.exe, writes
# %LOCALAPPDATA%\TaxiCheck, user PATH, and a Start Menu shortcut.
build-windows-installer: build-windows
	@mkdir -p cmd/install/payload
	cp dist/$(APP_NAME)-windows-amd64.exe cmd/install/payload/taxiprijs.exe
	cp .env.example cmd/install/payload/env.example
	cp LICENSE cmd/install/payload/LICENSE
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(INSTALLER_NAME) ./cmd/install
	@echo "Windows installer: dist/$(INSTALLER_NAME)"

build-all: build-linux build-macos build-windows build-windows-installer
	@echo "Builds complete in dist/"

# User-writable install: binary, source-repo marker, Omarchy QML plugin.
# No sudo. Safe to run from the TUI after a pull or F3.
install-user:
	@mkdir -p $(HOME)/.local/bin $(HOME)/.taxiprijs
	@cp $(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)
	@pwd > $(HOME)/.taxiprijs/source-repo
	@mkdir -p $(PLUGIN_DST)
	@cp extras/omarchy-plugin/* $(PLUGIN_DST)/
	@if [ -f "$(HOME)/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq --arg id $(PLUGIN_ID) 'if ([.bar.layout.center[]? , .bar.layout.left[]? , .bar.layout.right[]?] | any(.id == $$id)) then . else .bar.layout.right = ((.bar.layout.right // []) + [{"id": $$id}]) end' "$(HOME)/.config/omarchy/shell.json" > "$(HOME)/.config/omarchy/shell.json.tmp" && mv "$(HOME)/.config/omarchy/shell.json.tmp" "$(HOME)/.config/omarchy/shell.json"; \
	fi
	@if command -v omarchy >/dev/null 2>&1; then omarchy restart shell || true; \
	elif command -v omarchy-shell >/dev/null 2>&1; then omarchy-shell -q shell rescanPlugins || true; fi
	@echo "Installed: $(HOME)/.local/bin/$(APP_NAME)"
	@echo "Omarchy QML plugin: $(PLUGIN_DST)"

install-system:
	sudo install -Dm755 $(APP_NAME) /usr/local/bin/$(APP_NAME)
	sudo install -Dm644 extras/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop
	@echo "Installed: /usr/local/bin/$(APP_NAME)"
	@echo "Desktop entry: /usr/share/applications/$(APP_NAME).desktop"

# Always update the user binary + QML. System files only if passwordless sudo
# is available, so `make install` from the TUI never blocks on a password.
install: build install-user
	@sudo -n install -Dm755 $(APP_NAME) /usr/local/bin/$(APP_NAME) 2>/dev/null || true
	@sudo -n install -Dm644 extras/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop 2>/dev/null || true

uninstall:
	@sudo -n rm -f /usr/local/bin/$(APP_NAME) 2>/dev/null || true
	@sudo -n rm -f /usr/share/applications/$(APP_NAME).desktop 2>/dev/null || true
	@sudo -n rm -f /usr/local/share/man/man1/$(APP_NAME).1 2>/dev/null || true
	rm -f $$HOME/.local/bin/$(APP_NAME)
	rm -rf $$HOME/.$(APP_NAME)
	rm -rf $$HOME/.config/omarchy/plugins/jp.taxiprijs
	@if [ -f "$$HOME/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq '(.bar.layout.center // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.left // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.right // []) |= map(select(.id != "jp.taxiprijs"))' "$$HOME/.config/omarchy/shell.json" > "$$HOME/.config/omarchy/shell.json.tmp" && mv "$$HOME/.config/omarchy/shell.json.tmp" "$$HOME/.config/omarchy/shell.json"; \
	fi
	@echo "Uninstalled"

clean:
	rm -rf dist/ $(APP_NAME)
