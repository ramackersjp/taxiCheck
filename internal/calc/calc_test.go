package calc

import (
	"testing"
)

func TestCalculate1to4Passengers(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "1-4 passengers", BoardFee: 3.50, PerMinute: 0.50, WaitMinute: 0.50},
		{Name: "1-5 passengers", BoardFee: 5.00, PerMinute: 0.65, WaitMinute: 0.65},
	}

	input := FareInput{
		Minutes:    10,
		WaitTime:   2,
		Passengers: 3,
	}

	result := Calculate(input, groups)

	if result.Group != "1-4 passengers" {
		t.Errorf("expected group '1-4 passengers', got %q", result.Group)
	}

	if result.BaseFee != 3.50 {
		t.Errorf("expected base fee 3.50, got %.2f", result.BaseFee)
	}

	expectedTimeFee := 10 * 0.50
	if result.TimeFee != expectedTimeFee {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedWaitFee := 2 * 0.50
	if result.WaitFee != expectedWaitFee {
		t.Errorf("expected wait fee %.2f, got %.2f", expectedWaitFee, result.WaitFee)
	}

	expectedTotal := 3.50 + expectedTimeFee + expectedWaitFee
	if result.Total != expectedTotal {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestCalculate1to5Passengers(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "1-4 passengers", BoardFee: 3.50, PerMinute: 0.50, WaitMinute: 0.50},
		{Name: "1-5 passengers", BoardFee: 5.00, PerMinute: 0.65, WaitMinute: 0.65},
	}

	input := FareInput{
		Minutes:    20,
		WaitTime:   5,
		Passengers: 5,
	}

	result := Calculate(input, groups)

	if result.Group != "1-5 passengers" {
		t.Errorf("expected group '1-5 passengers', got %q", result.Group)
	}

	if result.BaseFee != 5.00 {
		t.Errorf("expected base fee 5.00, got %.2f", result.BaseFee)
	}

	expectedTimeFee := 20 * 0.65
	if result.TimeFee != expectedTimeFee {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedWaitFee := 5 * 0.65
	if result.WaitFee != expectedWaitFee {
		t.Errorf("expected wait fee %.2f, got %.2f", expectedWaitFee, result.WaitFee)
	}

	expectedTotal := 5.00 + expectedTimeFee + expectedWaitFee
	if result.Total != expectedTotal {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestCalculateDefaultGroup(t *testing.T) {
	groups := []PassengerGroup{}

	input := FareInput{
		Minutes:    5,
		WaitTime:   0,
		Passengers: 2,
	}

	result := Calculate(input, groups)

	if result.Group != "default" {
		t.Errorf("expected group 'default', got %q", result.Group)
	}

	if result.BaseFee != 3.50 {
		t.Errorf("expected base fee 3.50, got %.2f", result.BaseFee)
	}

	expectedTimeFee := 5 * 0.50
	if result.TimeFee != expectedTimeFee {
		t.Errorf("expected time fee %.2f, got %.2f", expectedTimeFee, result.TimeFee)
	}

	expectedTotal := 3.50 + expectedTimeFee
	if result.Total != expectedTotal {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, result.Total)
	}
}

func TestSelectGroup(t *testing.T) {
	groups := []PassengerGroup{
		{Name: "1-4 passengers", BoardFee: 3.50, PerMinute: 0.50, WaitMinute: 0.50},
		{Name: "1-5 passengers", BoardFee: 5.00, PerMinute: 0.65, WaitMinute: 0.65},
	}

	tests := []struct {
		passengers int
		wantGroup  string
	}{
		{1, "1-4 passengers"},
		{4, "1-4 passengers"},
		{5, "1-5 passengers"},
		{0, "1-4 passengers"},
		{6, "1-4 passengers"},
	}

	for _, tt := range tests {
		group := selectGroup(tt.passengers, groups)
		if group.Name != tt.wantGroup {
			t.Errorf("passengers=%d: expected group %q, got %q", tt.passengers, tt.wantGroup, group.Name)
		}
	}
}
