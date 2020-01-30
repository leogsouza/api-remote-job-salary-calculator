package main

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	data := []struct {
		tp   string
		amt  float64
		rt   float64
		resp *ResponseCalculateJSON
	}{
		{"annual", 70000.0, 4.2466225406, &ResponseCalculateJSON{70000.0, 5833.333333333333, 24771.964820166664, 18415.47524462083}},
	}

	for _, tt := range data {
		resp := calculateAnnual(tt.amt, tt.rt)
		if resp.AnnualSalary != tt.resp.AnnualSalary {
			t.Errorf("Annual salary expected: %f but got %f", tt.resp.AnnualSalary, resp.AnnualSalary)
		}
	}
}
