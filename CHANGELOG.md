# Changelog

All notable changes to TaxiCheck are documented here.

## [V2.0.1] - 2026-09-01

### Added
- **Snap packaging** (`taxiprijs`) — reproducible `snapcraft.yaml` (core24, strict confinement), app icon/desktop entry, and a CI workflow that builds the snap on pulls to `dev` and on release tags. The snap gets `network` and `home` plugs.
- **Flatpak packaging** (`dev.ramackers.TaxiCheck`) — flatpak manifest (freedesktop 24.08), desktop entry, icon, AppStream metadata, and a CI workflow that builds the flatpak bundle on pulls to `dev` and on release tags.
- **Docker image** (`ghcr.io/ramackersjp/taxicheck`) — multi-stage Dockerfile (Go build stage + slim runtime), `.dockerignore`, and a release workflow that builds and publishes the image to GitHub Container Registry with the version tag and `latest`.
- **Snap and Flatpak artifacts now attach to the GitHub release** — `release.yml` builds the Windows installer, snap, and flatpak on a published release and uploads all three to the release assets.
- **Docker section** in the README with install, version-tag, and local-build/run instructions.

### Changed
- Version bumped to **V2.0.1** across the repo (`Makefile`, snap manifest, flatpak AppStream metadata, man page, README, prompts, Omarchy plugin manifest, installer/binaries).
- The home screen's `1 ▸ Menu` label now dims the version number and uses colons on the header info labels.
- In-app copyright year and MIT license year updated to 2026.

### Fixed
- A desktop entry is now always installed without requiring `sudo` (user-level entry is always placed; system files only with passwordless sudo).
- The flatpak bundle build now creates the `dist` directory first, and the flatpak artifact is uploaded via a slash-free bundle path.

### Distribution
Available install forms for V2.0.1:

- **Windows** — `install-taxicheck-v2.0.1-windows.exe` on GitHub releases (no Go/Git/compiler needed).
- **Snap** — `taxiprijs` (from source or the Snap Store).
- **Flatpak** — `dev.ramackers.TaxiCheck` (from source or Flathub).
- **Docker** — `ghcr.io/ramackersjp/taxicheck:2.0.1` / `ghcr.io/ramackersjp/taxicheck:latest`.
- **From source** — `git clone` + `git checkout v2.0.1`, then `make install` or build with Go.