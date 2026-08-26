# TaxiCheck — Omarchy Plugin

A bar widget for Omarchy Quattro that launches the TaxiCheck TUI.

## Install (after `make install`)

```sh
omarchy plugin add https://github.com/ramackersjp/taxiprijs.git --path extras/omarchy-plugin
```

Or manually:

```sh
PLUGIN_DIR="$HOME/.config/omarchy/plugins/jp.taxiprijs"
mkdir -p "$PLUGIN_DIR"
cp extras/omarchy-plugin/* "$PLUGIN_DIR/"
omarchy-shell shell rescanPlugins
```

Then add to your bar layout in `~/.config/omarchy/shell.json`:

```json
{ "id": "jp.taxiprijs" }
```

## Remove

```sh
omarchy plugin remove jp.taxiprijs
```
