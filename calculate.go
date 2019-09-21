package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Brazilian taxes
const (
	AIRRF    = 0.275
	VLPERDEP = 189.59
	DIRRF    = 869.36
	INSS     = 642.34
)

// ResponseCalculateJSON is the response of the method calculate
type ResponseCalculateJSON struct {
	AnnualSalary     float64 `json:"annual_salary"`
	MonthlySalary    float64 `json:"monthly_salary"`
	ConvertedSalary  float64 `json:"converted_salary"`
	CalculatedSalary float64 `json:"calculated_salary"`
}

type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

func calculate(c *gin.Context) {
	fromCurrency := c.Param("from")
	toCurrency := c.Param("to")
	amount, err := strconv.ParseFloat(c.Param("amount"), 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	realRate, err := convertExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	response := ResponseCalculateJSON{}
	response.AnnualSalary = amount
	response.MonthlySalary = amount / 12
	response.ConvertedSalary = realRate * response.MonthlySalary
	response.CalculatedSalary = calculateBrazilianSalary(response.ConvertedSalary)
	c.JSON(http.StatusOK, response)
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
