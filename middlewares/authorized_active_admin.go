package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/models"
)

func AuthorizedActiveAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := c.MustGet("admin").(models.Admin)
		if admin.Status != "Active" || admin.User.Status != "Activated" {
			c.JSON(422, gin.H{
				"status": "failed",
				"errors": "Admin/user sedang berstatus tidak aktif.",
				"result": nil,
				"description": "Tidak dapat melanjutkan request karena Admin berstatus tidak aktif.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}