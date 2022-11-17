package controllers

import (
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	response := "Hello from dashboard"
	c.Data(200, "text/html; charset:utf-8", []byte(response))
}