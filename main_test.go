package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateHandler(t *testing.T) {

	url := "localhost:8080/salary/calculator?"
	data := []struct {
		urlParams  string
		statusCode int
		resp       interface{}
	}{
		{"type=daily&from=USD&to=BRL&amount=500", http.StatusOK, &ResponseCalculateJSON{120000, 10000, 42466.225406, 31243.81416935}},
		{"type=daly&from=USD&to=BRL&amount=500", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
	}

	for _, tt := range data {
		req, err := http.NewRequest("GET", url+tt.urlParams, nil)
		if err != nil {
			t.Fatalf("could not created request: %v", err)
		}
		rec := httptest.NewRecorder()

		calculateHandler(rec, req)

		res := rec.Result()
		if res.StatusCode != tt.statusCode {
			t.Errorf("expected %d; got %v", tt.statusCode, res.StatusCode)
		}
	}

}

func TestCalculate(t *testing.T) {
	data := []struct {
		tp   string
		fc   string
		tc   string
		amt  float64
		rt   float64
		resp *ResponseCalculateJSON
	}{
		{"annual", "USD", "BRL", 70000.0, 4.2466225406, &ResponseCalculateJSON{70000.0, 5833.333333333333, 24771.964820166664, 18415.47524462083}},
		{"monthly", "USD", "BRL", 6000, 4.2466225406, &ResponseCalculateJSON{72000, 6000, 25479.7352436, 18928.60880161}},
		{"daily", "USD", "BRL", 500, 4.2466225406, &ResponseCalculateJSON{120000, 10000, 42466.225406, 31243.81416935}},
		{"hourly", "USD", "BRL", 60, 4.2466225406, &ResponseCalculateJSON{115200, 9600, 40767.57638976, 30012.293632576002}},
	}

	for _, tt := range data {
		resp := calculate(tt.tp, tt.fc, tt.tc, tt.amt, tt.rt)
		if resp.AnnualSalary != tt.resp.AnnualSalary {
			t.Errorf("Annual salary expected: %f but got %f", tt.resp.AnnualSalary, resp.AnnualSalary)
		}

		if resp.MonthlySalary != tt.resp.MonthlySalary {
			t.Errorf("Monthly salary expected: %f but got %f", tt.resp.MonthlySalary, resp.MonthlySalary)
		}

		if resp.ConvertedSalary != tt.resp.ConvertedSalary {
			t.Errorf("Converted salary expected: %f but got %f", tt.resp.ConvertedSalary, resp.ConvertedSalary)
		}

		if resp.CalculatedSalary != tt.resp.CalculatedSalary {
			t.Errorf("Calculated salary expected: %f but got %f", tt.resp.CalculatedSalary, resp.CalculatedSalary)
		}
	}
}
