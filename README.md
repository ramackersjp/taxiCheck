# Taxiprijs

A modern, lightweight terminal user interface (TUI) for calculating Dutch taxi fares.

## Features

- Calculate taxi fares based on configurable rates
- Support for multiple passenger groups (1-4 and 1-5 passengers)
- Configurable pricing: board fee, per minute rate, waiting time rate
- TOML-based configuration that can be manually edited
- Settings screen to modify pricing without editing files
- Clean, professional taxi-inspired UI with yellow borders
- ASCII taxi logo

## Installation

### From Source

```bash
git clone https://github.com/jp/taxiprijs.git
cd taxiprijs
go build -o taxiprijs ./cmd/taxiprijs
```

### Prerequisites

- Go 1.21 or later

### Recommended Font

For the best experience, use [JetBrains Nerd Font](https://www.nerdfonts.com/). The application will work with any terminal font, but the styling is optimized for JetBrains Nerd Font.

## Usage

### Running the Application

```bash
./taxiprijs
```

### First Run

On first launch, the application will detect that no configuration exists and offer to perform initial setup. You can:

1. Enter pricing for each passenger group
2. Press Enter to save
3. Press Esc to cancel

### Main Menu

- **1** - Calculate Fare
- **2** - Settings
- **3** - Help/Manual
- **4** - Initial Setup
- **q** - Quit

### Calculating a Fare

1. Enter trip duration in minutes
2. Enter waiting time in minutes
3. Enter number of passengers (1-5)
4. Press Enter to calculate
5. View the fare breakdown and total

### Keyboard Controls

| Key | Action |
|-----|--------|
| 1-4 | Select menu option |
| Tab | Next input field |
| Shift+Tab | Previous input field |
| Enter | Submit/Save |
| Esc | Back/Cancel |
| q | Quit |

## Configuration

### Configuration File Location

```
~/.taxiprijs/config.toml
```

### TOML Format

```toml
[[passenger_groups]]
name = "1-4 passengers"
board_fee = 3.50
per_minute = 0.50
wait_minute = 0.50

[[passenger_groups]]
name = "1-5 passengers"
board_fee = 5.00
per_minute = 0.65
wait_minute = 0.65
```

### Manual Configuration

You can edit the TOML file directly with any text editor. Changes will be loaded automatically when you restart the application.

### Settings Screen

Access Settings from the main menu (option 2) to modify pricing without manually editing the TOML file. Changes are saved automatically.

## Passenger Categories

| Category | Description |
|----------|-------------|
| 1-4 passengers | Standard group size |
| 1-5 passengers | Larger group (e.g., van) |

## Pricing

- **Board Fee**: Initial charge when entering the taxi
- **Per Minute**: Cost per minute of driving
- **Wait Minute**: Cost per minute of waiting

## Manual

A Unix-style man page is included. To view:

```bash
man ./taxiprijs.1
```

Or install it:

```bash
cp taxiprijs.1 /usr/local/share/man/man1/
man taxiprijs
```

## Development

### Project Structure

```
taxiprijs/
├── cmd/
│   └── taxiprijs/
│       └── main.go          # Application entry point
├── internal/
│   ├── calc/
│   │   ├── calc.go          # Fare calculation engine
│   │   └── calc_test.go     # Unit tests
│   ├── config/
│   │   ├── config.go        # TOML configuration
│   │   └── config_test.go   # Unit tests
│   └── tui/
│       ├── model.go         # Bubble Tea TUI
│       ├── style.go         # Lip Gloss styling
│       └── logo.go          # ASCII taxi logo
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── taxiprijs.1              # Man page
└── prompt.md                # Project knowledge
```

### Running Tests

```bash
go test ./...
```

### Code Quality

```bash
go vet ./...
gofmt -w .
```

## Git Workflow

- Development happens on the `dev` branch
- Never commit directly to `master`
- Create feature branches from `dev`
- Test before committing

## License

MIT License - see [LICENSE](LICENSE) file.

## Uninstall

```bash
# Remove binary
rm ./taxiprijs

# Remove configuration (optional)
rm -rf ~/.taxiprijs

# Remove man page (if installed)
sudo rm /usr/local/share/man/man1/taxiprijs.1
```

## Future Features

The architecture is designed to support future integration with:

- Real-time routing APIs
- Traffic information
- Actual driving distance
- Estimated arrival times
