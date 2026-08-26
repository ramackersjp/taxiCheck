package calc

import (
	"testing"
)

func approxEqual(a, b, tolerance float64) bool {
	if a-b < 0 {
		b, a = a, b
	}
	return a-b < tolerance
}

func TestCalculateTaxiAuto(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "Taxi auto (max 4)", BoardFee: 4.31, PerKm: 3.17, PerMinute: 0.52, WaitMinute: 59.41},
		{Name: "Taxi bus (5-8)", BoardFee: 8.77, PerKm: 4.00, PerMinute: 0.59, WaitMinute: 59.41},
	}

	input := FareInput{
		DistanceKm:  15.5,
		DurationMin: 22,
		Passengers:  3,
	}

	result := Calculate(input, groups)

	if result.Group != "Taxi auto (max 4)" {
		t.Errorf("expected group 'Taxi auto (max 4)', got %q", result.Group)
	}

	if result.BaseFee != 4.31 {
		t.Errorf("expected base fee 4.31, got %.2f", result.BaseFee)
	}

	expectedKmFee := 15.5 * 3.17
	if !approxEqual(result.KmFee, expectedKmFee, 0.01) {
		t.Errorf("expected km fee %.2f, got %.2f", expectedKmFee, result.KmFee)
	}

	expectedTimeFee := 22 * 0.52
	if !approxEqual(result.TimeFee, expectedTimeFee, 0.01) {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedTotal := 4.31 + expectedKmFee + expectedTimeFee
	if !approxEqual(result.Total, expectedTotal, 0.01) {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestCalculateTaxiBus(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "Taxi auto (max 4)", BoardFee: 4.31, PerKm: 3.17, PerMinute: 0.52, WaitMinute: 59.41},
		{Name: "Taxi bus (5-8)", BoardFee: 8.77, PerKm: 4.00, PerMinute: 0.59, WaitMinute: 59.41},
	}

	input := FareInput{
		DistanceKm:  80.0,
		DurationMin: 55,
		Passengers:  6,
	}

	result := Calculate(input, groups)

	if result.Group != "Taxi bus (5-8)" {
		t.Errorf("expected group 'Taxi bus (5-8)', got %q", result.Group)
	}

	if result.BaseFee != 8.77 {
		t.Errorf("expected base fee 8.77, got %.2f", result.BaseFee)
	}

	expectedKmFee := 80.0 * 4.00
	if !approxEqual(result.KmFee, expectedKmFee, 0.01) {
		t.Errorf("expected km fee %.2f, got %.2f", expectedKmFee, result.KmFee)
	}

	expectedTimeFee := 55 * 0.59
	if !approxEqual(result.TimeFee, expectedTimeFee, 0.01) {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedTotal := 8.77 + expectedKmFee + expectedTimeFee
	if !approxEqual(result.Total, expectedTotal, 0.01) {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestCalculateDefaultGroup(t *testing.T) {
	groups := []PassengerGroup{}

	input := FareInput{
		DistanceKm:  10.0,
		DurationMin: 15,
		Passengers:  2,
	}

	result := Calculate(input, groups)

	if result.Group != "default" {
		t.Errorf("expected group 'default', got %q", result.Group)
	}

	if result.BaseFee != 4.31 {
		t.Errorf("expected base fee 4.31, got %.2f", result.BaseFee)
	}

	expectedKmFee := 10.0 * 3.17
	if !approxEqual(result.KmFee, expectedKmFee, 0.01) {
		t.Errorf("expected km fee %.2f, got %.2f", expectedKmFee, result.KmFee)
	}

	expectedTimeFee := 15 * 0.52
	if !approxEqual(result.TimeFee, expectedTimeFee, 0.01) {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedTotal := 4.31 + expectedKmFee + expectedTimeFee
	if !approxEqual(result.Total, expectedTotal, 0.01) {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestSelectGroup(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "Taxi auto (max 4)", BoardFee: 4.31, PerKm: 3.17, PerMinute: 0.52, WaitMinute: 59.41},
		{Name: "Taxi bus (5-8)", BoardFee: 8.77, PerKm: 4.00, PerMinute: 0.59, WaitMinute: 59.41},
	}

	tests := []struct {
		passengers int
		wantGroup  string
	}{
		{1, "Taxi auto (max 4)"},
		{4, "Taxi auto (max 4)"},
		{5, "Taxi bus (5-8)"},
		{8, "Taxi bus (5-8)"},
		{0, "Taxi auto (max 4)"},
		{9, "Taxi auto (max 4)"},
	}

	for _, tt := range tests {
		group := selectGroup(tt.passengers, groups)
		if group.Name != tt.wantGroup {
			t.Errorf("passengers=%d: expected group %q, got %q", tt.passengers, tt.wantGroup, group.Name)
		}
	}
}
