package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/controllers"
	"github.com/adeindriawan/itsfood-administration/utils"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/adeindriawan/itsfood-administration/middlewares"
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

	authorized := r.Group("/")
	authorized.Use(middlewares.Authorized())
	{
		authorized.GET("/dummy/authorized", controllers.DummyAuthorizedController)
		authorizedAdmin := authorized.Group("/")
		authorizedAdmin.Use(middlewares.AuthorizedAdmin())
		{
			authorizedAdmin.GET("/dummy/authorized/admin", controllers.DummyAuthorizedAdminController)
			authorizedActiveAdmin := authorizedAdmin.Group("/")
			authorizedActiveAdmin.Use(middlewares.AuthorizedActiveAdmin())
			{
				authorizedActiveAdmin.POST("/orders/:orderId/vendor/:vendorId/notify", controllers.NotifyAVendorForAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/menu/:menuId/change", controllers.ChangeMenuInAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/menu/:menuId/qty", controllers.ChangeQtyOfAMenuInAnOrder)
			}
		}
	}


	r.GET("/admin", controllers.Dashboard)
	r.POST("/auth/login", controllers.AdminLogin)
	r.POST("/auth/register", controllers.AdminRegister)
	r.POST("/auth/logout", controllers.Logout)

	log.Fatal(r.Run(":8090"))
}