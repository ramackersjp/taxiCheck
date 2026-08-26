package calc

import (
	"fmt"
	"time"
)

type PassengerGroup struct {
	Name       string  `toml:"name"`
	BoardFee   float64 `toml:"board_fee"`
	PerMinute  float64 `toml:"per_minute"`
	WaitMinute float64 `toml:"wait_minute"`
}

type FareInput struct {
	Minutes    float64
	WaitTime   float64
	Passengers int
}

type FareResult struct {
	BaseFee float64
	TimeFee float64
	WaitFee float64
	Total   float64
	Group   string
}

func Calculate(input FareInput, groups []PassengerGroup) FareResult {
	group := selectGroup(input.Passengers, groups)

	baseFee := group.BoardFee
	timeFee := input.Minutes * group.PerMinute
	waitFee := input.WaitTime * group.WaitMinute

	return FareResult{
		BaseFee: baseFee,
		TimeFee: timeFee,
		WaitFee: waitFee,
		Total:   baseFee + timeFee + waitFee,
		Group:   group.Name,
	}
}

func selectGroup(passengers int, groups []PassengerGroup) PassengerGroup {
	for _, g := range groups {
		switch g.Name {
		case "1-5 passengers":
			if passengers >= 1 && passengers <= 5 {
				return g
			}
		case "1-4 passengers":
			if passengers >= 1 && passengers <= 4 {
				return g
			}
		}
	}

	if len(groups) > 0 {
		return groups[0]
	}

	return PassengerGroup{
		Name:       "default",
		BoardFee:   3.50,
		PerMinute:  0.50,
		WaitMinute: 0.50,
	}
}

func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
