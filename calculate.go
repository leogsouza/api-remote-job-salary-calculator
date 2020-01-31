package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

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

// Render sets the status code to request when and error happen
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrNotFound corresponds to error response when resource not found
var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

// ErrInvalidRequest returns an Invalid Request Error Response
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

// ExchangeRateResponse represents the information coming from Exchange API
type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()

	typeParam := keys.Get("type")
	if typeParam == "" {
		errType := errors.New("Parameter 'type' is required")
		render.Render(w, r, ErrInvalidRequest(errType))
		return
	}

	if errList := validateListValues([]string{"annual", "monthly", "hourly", "daily"}, typeParam); errList != nil {
		render.Render(w, r, ErrInvalidRequest(errList))
		return
	}

	fromCurrency := keys.Get("from")
	if fromCurrency == "" {
		errFrom := errors.New("Parameter 'from' is required")
		render.Render(w, r, ErrInvalidRequest(errFrom))
		return
	}

	toCurrency := keys.Get("to")
	if toCurrency == "" {
		errTo := errors.New("Parameter 'to' is required")
		render.Render(w, r, ErrInvalidRequest(errTo))
		return
	}
	amountParam := keys.Get("amount")
	if amountParam == "" {
		errAmount := errors.New("Parameter 'amount' is required")
		render.Render(w, r, ErrInvalidRequest(errAmount))
		return
	}
	amount, err := strconv.ParseFloat(amountParam, 64)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	realRate, err := convertExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	response := calculate(typeParam, fromCurrency, toCurrency, amount, realRate)
	render.JSON(w, r, response)
}

func calculate(typeParam, fromCurrency, toCurrency string, amount, realRate float64) *ResponseCalculateJSON {

	switch typeParam {
	case "annual":
		return calculateAnnual(amount, realRate)
	case "monthly":
		return calculateMonthly(amount, realRate)
	case "daily":
		return calculateDaily(amount, realRate)
	case "hourly":
		return calculateHourly(amount, realRate)
	}

	return nil
}

func calculateAnnual(amount float64, realRate float64) *ResponseCalculateJSON {
	response := &ResponseCalculateJSON{}
	response.AnnualSalary = amount
	response.MonthlySalary = amount / 12
	response.ConvertedSalary = realRate * response.MonthlySalary
	response.CalculatedSalary = calculateBrazilianSalary(response.ConvertedSalary)
	return response
}

func calculateMonthly(amount float64, realRate float64) *ResponseCalculateJSON {
	return calculateAnnual(amount*12, realRate)
}

func calculateDaily(amount float64, realRate float64) *ResponseCalculateJSON {
	return calculateMonthly(amount*20, realRate)
}

func calculateHourly(amount, realRate float64) *ResponseCalculateJSON {
	return calculateDaily(amount*8, realRate)
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

func validateListValues(list []string, value string) error {
	for _, item := range list {
		if item == value {
			return nil
		}
	}
	return fmt.Errorf("%s is not a valid value", value)
}
