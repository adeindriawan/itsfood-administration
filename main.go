package main

import (
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/controllers"
)

func main() {
	r := gin.Default()

	r.GET("/", func (c *gin.Context) {
		response := "Hello"
		fmt.Println("Hello world!")
		c.Data(200, "text/html; charset: utf-8", []byte(response))
	})
	r.GET("/admin", controllers.Dashboard)

	log.Fatal(r.Run())
}