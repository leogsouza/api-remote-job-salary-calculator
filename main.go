package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "I'm online"})
	})

	r.GET("/calculate/from/:from/to/:to/amount/:amount", calculate)

	r.Run(":80")
}
