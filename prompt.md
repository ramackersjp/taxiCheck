# Taxiprijs Project Knowledge

## Project Purpose
Dutch taxi fare calculator TUI application built with Go and Bubble Tea.

## Current State
- Initial implementation complete
- All core features implemented
- Tests passing
- Documentation complete

## Technology Stack
- Go 1.21+
- Bubble Tea (TUI framework)
- Lip Gloss (styling)
- go-toml/v2 (TOML configuration)

## Architecture
```
taxiprijs/
├── cmd/taxiprijs/main.go          # Entry point
├── internal/
│   ├── calc/calc.go               # Fare calculation engine
│   ├── config/config.go           # TOML configuration
│   └── tui/                       # Bubble Tea TUI
│       ├── model.go               # Main model with all screens
│       ├── style.go               # Lip Gloss styling
│       └── logo.go                # ASCII taxi logo
├── go.mod
├── LICENSE
├── README.md
├── taxiprijs.1                    # Man page
└── prompt.md                      # This file
```

## Key Design Decisions
1. Separated calculation engine from TUI for future API integration
2. TOML for human-readable configuration
3. Yellow taxi-inspired visual theme
4. Minimal dependencies (Bubble Tea, Lip Gloss, go-toml only)

## Configuration Model
- Stored in `~/.taxiprijs/config.toml`
- Supports two passenger groups: "1-4 passengers" and "1-5 passengers"
- Each group has: board_fee, per_minute, wait_minute rates
- Validated on load with clear error messages

## TUI Screens
1. Main Menu - Navigation hub
2. Calculate Fare - Input trip details and calculate price
3. Settings - Modify pricing rates
4. Help/Manual - Keyboard controls and documentation
5. Initial Setup - First-run configuration

## Fare Calculation
- Base fee + (minutes × per_minute) + (wait_minutes × wait_minute)
- Automatically selects passenger group based on passenger count
- Default group used if no groups configured

## Testing Requirements
- Unit tests for calc package
- Unit tests for config package
- Run with: `go test ./...`
- Vet with: `go vet ./...`
- Format with: `gofmt -w .`

## Git Workflow
- Development on `dev` branch
- Never commit directly to `master`
- Create feature branches from `dev`
- Test before committing

## Dependencies
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/bubbles
- github.com/pelletier/go-toml/v2

## Known Limitations
- No real-time traffic data
- No actual distance calculation
- No route information
- Manual passenger count required

## Future API Direction
- Google Maps integration for actual distances
- Traffic APIs for real-time delays
- Route optimization
- Estimated arrival times

## Important Constraints
- No external APIs in initial version
- Must work entirely locally
- No telemetry or network access
- Configuration must be human-editable

## Current Features
- Fare calculation for 1-5 passengers
- Configurable pricing rates
- TOML configuration with validation
- Settings screen for rate modification
- Initial setup wizard
- Help/Manual screen
- Yellow taxi-inspired UI
- ASCII taxi logo
- Keyboard-only navigation
