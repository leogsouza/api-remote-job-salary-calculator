package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ALIQ_IRRF   = 0.275
	VL_PER_DEP  = 189.59
	DEDUCT_IRRF = 869.36
	INSS        = 642.34
	DOLLAR      = 3.78
)

// ResponseCalculateJSON is the response of the method calculate
type ResponseCalculateJSON struct {
	AnnualSalary     float64 `json:"annual_salary"`
	MonthlySalary    float64 `json:"monthly_salary"`
	ConvertedSalary  float64 `json:"converted_salary"`
	CalculatedSalary float64 `json:"calculated_salary"`
}

func calculate(c *gin.Context) {
	// fromCurrency := c.Param("from")
	// toCurrency := c.Param("to")
	amount, err := strconv.ParseFloat(c.Param("amount"), 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}

	response := ResponseCalculateJSON{}
	response.AnnualSalary = amount
	response.MonthlySalary = amount / 12
	response.ConvertedSalary = DOLLAR * response.MonthlySalary
	response.CalculatedSalary = calculateBrazilianSalary(response.ConvertedSalary)
	c.JSON(http.StatusOK, response)
}

func calculateBrazilianSalary(amount float64) float64 {
	irrf := (amount-VL_PER_DEP-INSS)*ALIQ_IRRF - DEDUCT_IRRF

	return amount - INSS - irrf
}
