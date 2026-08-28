package calc

import (
	"strings"
)

type PassengerGroup struct {
	Name       string  `toml:"name"`
	BoardFee   float64 `toml:"board_fee"`
	PerKm      float64 `toml:"per_km"`
	PerMinute  float64 `toml:"per_minute"`
	WaitMinute float64 `toml:"wait_minute"`
}

type FareInput struct {
	DistanceKm  float64
	DurationMin float64
	Passengers  int
}

type FareResult struct {
	BaseFee float64
	KmFee   float64
	TimeFee float64
	Total   float64
	Group   string
}

func Calculate(input FareInput, groups []PassengerGroup) FareResult {
	group := selectGroup(input.Passengers, groups)

	baseFee := group.BoardFee
	kmFee := input.DistanceKm * group.PerKm
	timeFee := input.DurationMin * group.PerMinute

	return FareResult{
		BaseFee: baseFee,
		KmFee:   kmFee,
		TimeFee: timeFee,
		Total:   baseFee + kmFee + timeFee,
		Group:   group.Name,
	}
}

func selectGroup(passengers int, groups []PassengerGroup) PassengerGroup {
	for _, g := range groups {
		name := strings.ToLower(g.Name)
		switch {
		case strings.Contains(name, "max 4") || strings.Contains(name, "max. 4"):
			if passengers >= 1 && passengers <= 4 {
				return g
			}
		case strings.Contains(name, "5-8"):
			if passengers >= 5 && passengers <= 8 {
				return g
			}
		}
	}

	if len(groups) > 0 {
		return groups[0]
	}

	return PassengerGroup{
		Name:       "default",
		BoardFee:   4.31,
		PerKm:      3.17,
		PerMinute:  0.52,
		WaitMinute: 59.41,
	}
}
