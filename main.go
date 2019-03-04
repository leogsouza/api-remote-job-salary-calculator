package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "I'm online"})
	})

	r.GET("/calculate/from/:from/to/:to/amount/:amount", calculate)

	r.Run(fmt.Sprintf(":%s", port))
}
