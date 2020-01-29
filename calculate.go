package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// Brazilian taxes
const (
	AIRRF    = 0.275
	VLPERDEP = 189.59
	DIRRF    = 869.36
	INSS     = 642.34
)

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

// ResponseCalculateJSON is the response of the method calculate
type ResponseCalculateJSON struct {
	AnnualSalary     float64 `json:"annual_salary"`
	MonthlySalary    float64 `json:"monthly_salary"`
	ConvertedSalary  float64 `json:"converted_salary"`
	CalculatedSalary float64 `json:"calculated_salary"`
}

func (rd *ResponseCalculateJSON) Render(w http.ResponseWriter, r *http.Request) error {

	return nil
}

type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

func calculate(w http.ResponseWriter, r *http.Request) {
	fromCurrency := chi.URLParam(r, "from")
	toCurrency := chi.URLParam(r, "to")
	amount, err := strconv.ParseFloat(chi.URLParam(r, "amount"), 64)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	realRate, err := convertExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	response := &ResponseCalculateJSON{}
	response.AnnualSalary = amount
	response.MonthlySalary = amount / 12
	response.ConvertedSalary = realRate * response.MonthlySalary
	response.CalculatedSalary = calculateBrazilianSalary(response.ConvertedSalary)
	render.Render(w, r, response)
}

func calculateBrazilianSalary(amount float64) float64 {
	irrf := (amount-VLPERDEP-INSS)*AIRRF - DIRRF

	return amount - INSS - irrf
}

func convertExchangeRate(from string, to string) (float64, error) {

	url := fmt.Sprintf("https://api.exchangeratesapi.io/latest?base=%s&symbols=%s", from, to)
	response, err := http.Get(url)

	if err != nil {
		return 0, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	var exchange ExchangeRateResponse
	json.Unmarshal(responseData, &exchange)

	return exchange.Rates["BRL"], nil

}
