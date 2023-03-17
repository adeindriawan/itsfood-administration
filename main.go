package main

import (
	"log"
	"os"

	"github.com/adeindriawan/itsfood-administration/controllers"
	"github.com/adeindriawan/itsfood-administration/middlewares"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/adeindriawan/itsfood-administration/utils"
	"github.com/gin-gonic/gin"
)

func init() {
	utils.LoadEnvVars()
	services.InitRedis()
	services.InitMySQL()
}

func main() {
	r := gin.Default()
	r.Use(middlewares.CORS())

	r.GET("/", func(c *gin.Context) {
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
				authorizedActiveAdmin.GET("/orders", controllers.GetOrders)

				authorizedActiveAdmin.GET("/customers", controllers.GetCustomers)

				authorizedActiveAdmin.GET("/units", controllers.GetUnits)

				authorizedActiveAdmin.GET("/orders/:id", controllers.GetOrder)
				authorizedActiveAdmin.POST("/orders/:orderId/vendor/:vendorId/notify", controllers.NotifyAVendorForAnOrder)
				authorizedActiveAdmin.GET("/orders/:id/vendors", controllers.GetVendorsInAnOrder)

				authorizedActiveAdmin.POST("/order-details/:orderDetailId/menu/:menuId/change", controllers.ChangeMenuInAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/qty", controllers.ChangeQtyOfAMenuInAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/note", controllers.ChangeNoteOfAMenuInAnOrder)
				authorizedActiveAdmin.PATCH("/order-details/:orderDetailId/status", controllers.ChangeStatusOfAMenuInAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/cost", controllers.AddCostToAnOrder)
				authorizedActiveAdmin.POST("/order-details/:orderDetailId/discount", controllers.AddDiscountToAnOrder)
			}
		}
	}

	r.GET("/admin", controllers.Dashboard)
	r.POST("/auth/login", controllers.AdminLogin)
	r.POST("/auth/register", controllers.AdminRegister)
	r.POST("/auth/logout", controllers.Logout)
	r.POST("/token/refresh", controllers.Refresh)

	log.Fatal(r.Run(":" + os.Getenv("PORT")))
}
