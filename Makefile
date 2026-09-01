APP_NAME := taxiprijs
VERSION  := 2.0.1
GOFLAGS  := -trimpath
LDFLAGS  := -s -w -X main.version=$(VERSION)
INSTALLER_NAME := install-taxicheck-v$(VERSION)-windows.exe
PLUGIN_ID := jp.taxiprijs
PLUGIN_DST := $(HOME)/.config/omarchy/plugins/$(PLUGIN_ID)
FLATPAK_ID := dev.ramackers.TaxiCheck
ICON_SRC := extras/$(APP_NAME).svg

# Cross-compile and packaging targets
.PHONY: build build-linux build-macos build-windows build-windows-installer build-snap build-flatpak install install-user install-system uninstall clean

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

# Snap package (requires snapcraft). Output lands in dist/.
build-snap:
	@command -v snapcraft >/dev/null 2>&1 || { echo "snapcraft is required (see https://snapcraft.io/docs/installing-snapcraft)"; exit 1; }
	@mkdir -p dist
	@rm -f $(APP_NAME)_$(VERSION)_*.snap
	snapcraft --destructive-mode
	mv -f $(APP_NAME)_$(VERSION)_*.snap dist/

# Flatpak bundle (requires flatpak + flatpak-builder). Output lands in dist/.
# Run `flatpak install flathub org.freedesktop.Platform//24.08 org.freedesktop.Sdk//24.08`
# first if the runtimes aren't installed yet.
build-flatpak:
	@command -v flatpak-builder >/dev/null 2>&1 || { echo "flatpak-builder is required"; exit 1; }
	@mkdir -p dist
	rm -rf flatpak/repo flatpak/build-dir
	flatpak-builder --user --force-clean --disable-rofiles-fuse \
		--repo=flatpak/repo flatpak/build-dir flatpak/$(FLATPAK_ID).yaml
	flatpak remote-add --user --no-gpg-verify local-build flatpak/repo 2>/dev/null || true
	flatpak install --user -y local-build $(FLATPAK_ID)
	flatpak build-bundle $$HOME/.local/share/flatpak/repo \
		dist/$(APP_NAME)-flatpak-$(VERSION).flatpak $(FLATPAK_ID)
	@echo "Flatpak bundle: dist/$(APP_NAME)-flatpak-$(VERSION).flatpak"

# User-writable install: binary, source-repo marker, desktop entry + icon,
# Omarchy QML plugin. No sudo, so the desktop entry always lands even without
# passwordless sudo.
install-user:
	@mkdir -p $(HOME)/.local/bin $(HOME)/.taxiprijs
	@cp $(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)
	@pwd > $(HOME)/.taxiprijs/source-repo
	@install -Dm644 extras/$(APP_NAME).desktop $(HOME)/.local/share/applications/$(APP_NAME).desktop
	@install -Dm644 $(ICON_SRC) $(HOME)/.local/share/icons/hicolor/scalable/apps/$(APP_NAME).svg
	@mkdir -p $(PLUGIN_DST)
	@cp extras/omarchy-plugin/* $(PLUGIN_DST)/
	@if [ -f "$(HOME)/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq --arg id $(PLUGIN_ID) 'if ([.bar.layout.center[]? , .bar.layout.left[]? , .bar.layout.right[]?] | any(.id == $$id)) then . else .bar.layout.right = ((.bar.layout.right // []) + [{"id": $$id}]) end' "$(HOME)/.config/omarchy/shell.json" > "$(HOME)/.config/omarchy/shell.json.tmp" && mv "$(HOME)/.config/omarchy/shell.json.tmp" "$(HOME)/.config/omarchy/shell.json"; \
	fi
	@if command -v omarchy >/dev/null 2>&1; then omarchy restart shell || true; \
	elif command -v omarchy-shell >/dev/null 2>&1; then omarchy-shell -q shell rescanPlugins || true; fi
	@echo "Installed: $(HOME)/.local/bin/$(APP_NAME)"
	@echo "Desktop entry: $(HOME)/.local/share/applications/$(APP_NAME).desktop"
	@echo "Icon: $(HOME)/.local/share/icons/hicolor/scalable/apps/$(APP_NAME).svg"
	@echo "Omarchy QML plugin: $(PLUGIN_DST)"

install-system:
	sudo install -Dm755 $(APP_NAME) /usr/local/bin/$(APP_NAME)
	sudo install -Dm644 extras/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop
	sudo install -Dm644 $(ICON_SRC) /usr/share/icons/hicolor/scalable/apps/$(APP_NAME).svg
	@command -v gtk-update-icon-cache >/dev/null 2>&1 && sudo gtk-update-icon-cache -f /usr/share/icons/hicolor || true
	@command -v update-desktop-database >/dev/null 2>&1 && sudo update-desktop-database /usr/share/applications || true
	@echo "Installed: /usr/local/bin/$(APP_NAME)"
	@echo "Desktop entry: /usr/share/applications/$(APP_NAME).desktop"
	@echo "Icon: /usr/share/icons/hicolor/scalable/apps/$(APP_NAME).svg"

# Always update the user binary + user desktop entry + icon + QML. System files
# only if passwordless sudo is available, so `make install` from the TUI never
# blocks on a password (the user-level desktop entry still lands either way).
install: build install-user
	@sudo -n install -Dm755 $(APP_NAME) /usr/local/bin/$(APP_NAME) 2>/dev/null || true
	@sudo -n install -Dm644 extras/$(APP_NAME).desktop /usr/share/applications/$(APP_NAME).desktop 2>/dev/null || true
	@sudo -n install -Dm644 $(ICON_SRC) /usr/share/icons/hicolor/scalable/apps/$(APP_NAME).svg 2>/dev/null || true

uninstall:
	@sudo -n rm -f /usr/local/bin/$(APP_NAME) 2>/dev/null || true
	@sudo -n rm -f /usr/share/applications/$(APP_NAME).desktop 2>/dev/null || true
	@sudo -n rm -f /usr/share/icons/hicolor/scalable/apps/$(APP_NAME).svg 2>/dev/null || true
	@sudo -n rm -f /usr/local/share/man/man1/$(APP_NAME).1 2>/dev/null || true
	rm -f $$HOME/.local/bin/$(APP_NAME)
	rm -f $$HOME/.local/share/applications/$(APP_NAME).desktop
	rm -f $$HOME/.local/share/icons/hicolor/scalable/apps/$(APP_NAME).svg
	rm -rf $$HOME/.$(APP_NAME)
	rm -rf $$HOME/.config/omarchy/plugins/jp.taxiprijs
	@if [ -f "$$HOME/.config/omarchy/shell.json" ] && command -v jq >/dev/null 2>&1; then \
		jq '(.bar.layout.center // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.left // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.right // []) |= map(select(.id != "jp.taxiprijs"))' "$$HOME/.config/omarchy/shell.json" > "$$HOME/.config/omarchy/shell.json.tmp" && mv "$$HOME/.config/omarchy/shell.json.tmp" "$$HOME/.config/omarchy/shell.json"; \
	fi
	@echo "Uninstalled"

clean:
	rm -rf dist/ $(APP_NAME)
