package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/utils"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/adeindriawan/itsfood-administration/models"
)

func AuthorizedAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.AuthCheck(c)
		if err != nil {
			c.JSON(403, gin.H{
				"status": "failed",
				"errors": err.Error(),
				"results": nil,
				"description": "Gagal mengecek user ID pada request ini.",
			})
			c.Abort()
			return
		}
		
		var admin models.Admin
			if err := services.DB.Preload("User").Where("user_id = ?", userId).First(&admin).Error; err != nil {
				c.JSON(404, gin.H{
					"status": "failed",
					"errors": err.Error(),
					"result": userId,
					"description": "Gagal mengambil data admin dengan user ID yang dimaksud.",
				})
				c.Abort()
				return
			}
			c.Set("admin", admin) // add customer object to the context so it can be brought to next middleware
			c.Next()
	}
}