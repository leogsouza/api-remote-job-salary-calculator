package calculate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/logging"
	firebase "firebase.google.com/go"
	"github.com/go-chi/render"
)

const projectID = "remote-job-salary-calculator"

type taxesConfig struct {
	AliqIRRF   float64 `json:"ALIQ_IRRF"`
	DeductIRRF float64 `json:"DEDUCT_IRRF"`
	Inss       float64 `json:"INSS"`
	VlPerDep   float64 `json:"VL_PER_DEP"`
}

type secretsConfig struct {
	ApiURL string `json:"API_URL"`
	ApiKey string `json:"API_KEY"`
}

type config struct {
	taxes   taxesConfig
	secrets secretsConfig
}

var cfg config

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

var logger *log.Logger

func init() {
	ctx := context.Background()

	// Creates a client.
	logClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client logging: %v", err)
	}
	//defer logClient.Close()

	logName := "my-log"

	logger = logClient.Logger(logName).StandardLogger(logging.Info)

	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)

	logger.Println("Firebase connected", app)
	if err != nil {
		logger.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		logger.Fatalln(err)
	}

	secretsDoc, err := client.Collection("config").Doc("secrets").Get(ctx)
	if err != nil {
		logger.Fatalln(err)
	}

	taxesDoc, err := client.Collection("config").Doc("taxes").Get(ctx)
	if err != nil {
		logger.Fatalln(err)
	}

	taxesCfg := taxesConfig{
		AliqIRRF:   taxesDoc.Data()["ALIQ_IRRF"].(float64),
		DeductIRRF: taxesDoc.Data()["DEDUCT_IRRF"].(float64),
		Inss:       taxesDoc.Data()["INSS"].(float64),
		VlPerDep:   taxesDoc.Data()["VL_PER_DEP"].(float64),
	}

	secretsCfg := secretsConfig{
		ApiURL: secretsDoc.Data()["API_URL"].(string),
		ApiKey: secretsDoc.Data()["API_KEY"].(string),
	}

	cfg = config{
		taxes:   taxesCfg,
		secrets: secretsCfg,
	}

	defer client.Close()

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

type MetaResponse struct {
	Code       int    `json:"code"`
	Disclaimer string `json:"disclaimer"`
}

// ExchangeRateResponse represents the information coming from Exchange API
type APIExchangeRateResponse struct {
	Meta     MetaResponse `json:"meta"`
	Response ExchangeRateResponse
}

// ExchangeRateResponse represents the information coming from Exchange API
type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
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

	hoursParam := keys.Get("hours")
	var hoursPerDay float64 = 8
	if hoursParam != "" {
		hoursPerDay, err = strconv.ParseFloat(hoursParam, 64)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
	}

	realRate, err := convertExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	response := calculate(typeParam, fromCurrency, toCurrency, amount, realRate, hoursPerDay)
	render.JSON(w, r, response)
}

func calculate(typeParam, fromCurrency, toCurrency string, amount, realRate, hoursPerDay float64) *ResponseCalculateJSON {

	switch typeParam {
	case "annual":
		return calculateAnnual(amount, realRate)
	case "monthly":
		return calculateMonthly(amount, realRate)
	case "daily":
		return calculateDaily(amount, realRate)
	case "hourly":
		return calculateHourly(amount, realRate, hoursPerDay)
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

func calculateHourly(amount, realRate, hoursPerDay float64) *ResponseCalculateJSON {
	return calculateDaily(amount*hoursPerDay, realRate)
}

func calculateBrazilianSalary(amount float64) float64 {

	irrf := (amount-cfg.taxes.VlPerDep-cfg.taxes.Inss)*cfg.taxes.AliqIRRF - cfg.taxes.DeductIRRF

	return amount - cfg.taxes.Inss - irrf
}

func convertExchangeRate(from string, to string) (float64, error) {

	apiURL := cfg.secrets.ApiURL
	apiKey := cfg.secrets.ApiKey

	url := fmt.Sprintf("%s?api_key=%s&base=%s&symbols=%s", apiURL, apiKey, from, to)

	response, err := http.Get(url)

	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusOK {
		return 0, errors.New("Could not request the rate")
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	var exchange APIExchangeRateResponse
	json.Unmarshal(responseData, &exchange)

	return exchange.Response.Rates[to], nil

}

func validateListValues(list []string, value string) error {
	for _, item := range list {
		if item == value {
			return nil
		}
	}
	return fmt.Errorf("%s is not a valid value", value)
}
