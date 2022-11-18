package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/controllers"
	"github.com/adeindriawan/itsfood-administration/utils"
	"github.com/adeindriawan/itsfood-administration/services"
)

func init() {
	utils.LoadEnvVars()
	services.InitRedis()
	services.InitMySQL()
}

func main() {
	r := gin.Default()

	r.GET("/", func (c *gin.Context) {
		response := "This is Itsfood Administration Service API Homepage."
		c.Data(200, "text/html; charset: utf-8", []byte(response))
	})
	r.GET("/admin", controllers.Dashboard)
	r.POST("/auth/admin/login", controllers.AdminLogin)

	log.Fatal(r.Run(":8090"))
}