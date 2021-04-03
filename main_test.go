package main

import (
	"encoding/json"
	"fmt"
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
		{"from=USD&to=BRL&amount=500", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
		{"type=daily&to=BRL&amount=500", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
		{"type=daily&from=USD&amount=500", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
		{"type=daily&from=USD&to=BRL", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
		{"type=daily&from=USD&to=BRL&amount=abcd", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
		{"type=daily&from=USD&to=ABCD&amount=500", http.StatusBadRequest, &ErrResponse{StatusText: "Invalid request."}},
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
	tt := []struct {
		nm   string
		tp   string
		fc   string
		tc   string
		amt  float64
		rt   float64
		hpd  float64
		resp *ResponseCalculateJSON
	}{
		{"Annual Salary", "annual", "USD", "BRL", 70000.0, 4.2466225406, 8, &ResponseCalculateJSON{70000.0, 5833.333333333333, 24771.964820166664, 18415.47524462083}},
		{"Monthly Salary", "monthly", "USD", "BRL", 6000, 4.2466225406, 8, &ResponseCalculateJSON{72000, 6000, 25479.7352436, 18928.60880161}},
		{"Daily Salary", "daily", "USD", "BRL", 500, 4.2466225406, 8, &ResponseCalculateJSON{120000, 10000, 42466.225406, 31243.81416935}},
		{"Hourly Salary", "hourly", "USD", "BRL", 60, 4.2466225406, 8, &ResponseCalculateJSON{115200, 9600, 40767.57638976, 30012.293632576002}},
	}

	for _, tc := range tt {
		t.Run(tc.nm, func(t *testing.T) {
			resp := calculate(tc.tp, tc.fc, tc.tc, tc.amt, tc.rt, tc.hpd)
			if resp.AnnualSalary != tc.resp.AnnualSalary {
				t.Errorf("Annual salary expected: %f but got %f", tc.resp.AnnualSalary, resp.AnnualSalary)
			}

			if resp.MonthlySalary != tc.resp.MonthlySalary {
				t.Errorf("Monthly salary expected: %f but got %f", tc.resp.MonthlySalary, resp.MonthlySalary)
			}

			if resp.ConvertedSalary != tc.resp.ConvertedSalary {
				t.Errorf("Converted salary expected: %f but got %f", tc.resp.ConvertedSalary, resp.ConvertedSalary)
			}

			if resp.CalculatedSalary != tc.resp.CalculatedSalary {
				t.Errorf("Calculated salary expected: %f but got %f", tc.resp.CalculatedSalary, resp.CalculatedSalary)
			}
		})

	}
}

func TestRouting(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/salary/calculator?type=daily&from=USD&to=BRL&amount=500", srv.URL))
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	var respUser ResponseCalculateJSON
	err = json.NewDecoder(res.Body).Decode(&respUser)
	if err != nil {
		t.Fatalf("could not decode json")
	}
}
