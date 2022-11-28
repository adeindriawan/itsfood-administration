package controllers

import (
	"github.com/gin-gonic/gin"
)

func DummyAuthorizedController(c *gin.Context) {
	c.Data(200, "text/html; charset: utf-8", []byte("this is dummy authorized controller."))
}

func DummyAuthorizedAdminController(c *gin.Context) {
	c.Data(200, "text/html; charset: utf-8", []byte("this is dummy authorized admin controller."))
}

func NotifyAVendorForAnOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	vendorId := c.Param("vendorId")
	c.Data(200, "text/html; charset: utf-8", []byte("Do you want to notify the vendor "+ vendorId +" in order #"+ orderId +"?"))
}