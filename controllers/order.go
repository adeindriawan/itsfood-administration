package controllers

import (
	"time"
	"regexp"
	"errors"
	"strings"
	"strconv"
	"net/url"
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/models"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/adeindriawan/itsfood-administration/utils"
)

func DummyAuthorizedController(c *gin.Context) {
	c.Data(200, "text/html; charset: utf-8", []byte("this is dummy authorized controller."))
}

func DummyAuthorizedAdminController(c *gin.Context) {
	c.Data(200, "text/html; charset: utf-8", []byte("this is dummy authorized admin controller."))
}

type OrderResult struct {
	ID uint64							`json:"id"`
	OrderedFor time.Time	`json:"ordered_for"`
	OrderedTo string			`json:"ordered_to"`
	CustomerName string 	`json:"customer_name"`
	CustomerPhone string 	`json:"customer_phone"`
	CustomerUnit string		`json:"customer_unit"`
	CreatedAt time.Time		`json:"created_at"`
}

type OrderDetailResult struct {
	ID uint64						`json:"id"`
	MenuName string 		`json:"menu_name"`
	MenuQty uint 				`json:"menu_qty"`
	VendorName string		`json:"vendor_name"`
	VendorPhone string	`json:"vendor_phone"`
	Note string					`json:"note"`
	Status string				`json:"status"`
}

func NotifyAVendorForAnOrder(c *gin.Context) {
	var orderDetailModels []models.OrderDetail
	var order OrderResult
	var orderDetails []OrderDetailResult
	orderId := c.Param("orderId")
	vendorId := c.Param("vendorId")
	adminContext := c.MustGet("admin").(models.Admin)

	orderQuery := services.DB.Table("orders o").
		Select(`o.id AS ID, o.ordered_for AS OrderedFor, o.ordered_to AS OrderedTo,
			u.name AS CustomerName, u.phone AS CustomerPhone, n.name AS CustomerUnit,
			o.created_at AS CreatedAt
		`).
		Joins("JOIN customers c ON o.ordered_by = c.id").
		Joins("JOIN users u ON u.id = c.user_id").
		Joins("JOIN units n ON n.id = c.unit_id").
		Where("o.id = ?", orderId).
		Scan(&order)

	if orderQuery.Error != nil {
		c.JSON(200, gin.H{
			"status": "failed",
			"errors": orderQuery.Error.Error(),
			"description": "Gagal mengeksekusi query Order.",
			"result": nil,
		})
		return
	}

	orderDetailQuery := services.DB.Table("order_details od").
		Select(`od.id AS ID, m.name AS MenuName, od.qty AS MenuQty, u.name AS VendorName,
			v.phone AS VendorPhone, od.note AS Note, od.status AS Status
		`).
		Joins("JOIN orders o ON o.id = od.order_id").
		Joins("JOIN menus m ON m.id = od.menu_id").
		Joins("JOIN vendors v ON v.id = m.vendor_id").
		Joins("JOIN users u ON u.id = v.user_id").
		Where("o.id = ?", orderId).
		Where("v.id = ?", vendorId).
		Scan(&orderDetails)

	if orderDetailQuery.Error != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": orderDetailQuery.Error.Error(),
			"result": nil,
			"description": "Gagal mengeksekusi query Order details.",
		})
		return
	}

	if orderDetails[0].Status == "Sent" {
		c.JSON(200, gin.H{
			"status": "failed",
			"errors": "Tidak dapat mengirim notifikasi order baru ke vendor ini.",
			"result": orderDetails[0],
			"description": "Order details ada yang sudah berstatus Sent.",
		})
		return
	}

	vendorPhone := orderDetails[0].VendorPhone
	vendorPhoneNumber, err := sanitizePhoneNumber(vendorPhone)
	if err != nil {
		c.JSON(200, gin.H{
			"status": "failed",
			"result": nil,
			"errors": err.Error(),
			"description": "Gagal mengolah data nomor telepon vendor.",
		})
		return
	}

	orderedAt := utils.ConvertDateToPhrase(order.CreatedAt, true)
	orderedFor := utils.ConvertDateToPhrase(order.OrderedFor, true)
	var details = ""
	for _, item := range orderDetails {
		menuQty := strconv.Itoa(int(item.MenuQty))
		details += item.MenuName + " " + menuQty + " porsi. Catatan: " + item.Note
		models.UpdateOrderDetail(map[string]interface{}{"id": item.ID}, map[string]interface{}{"status": "Sent", "updated_at": time.Now(), "created_by": adminContext.User.Name})
	}

	var orderDetailSentStatus = []string{}
	services.DB.Where("order_id = ?", orderId).Find(&orderDetailModels)
	for _, item := range orderDetailModels {
		if item.Status == "Sent" {
			orderDetailSentStatus = append(orderDetailSentStatus, "Sent")
		}
	}

	var status string
	if len(orderDetailSentStatus) == len(orderDetailModels) {
		status = "ForwardedEntirely"
	} else {
		status = "ForwardedPartially"
	}
	
	models.UpdateOrder(map[string]interface{}{"id": orderId}, map[string]interface{}{"status": status, "updated_at": time.Now(), "created_by": adminContext.User.Name})

	message := "Ada order untuk " + orderDetails[0].VendorName + " dengan ID #" + orderId + " dari " + order.CustomerName + " di " + order.CustomerUnit
	message += " pada " + orderedAt + " untuk diantar pada " + orderedFor + " dengan rincian:\n"
	message += details

	whatsappAPI := "https://api.whatsapp.com/send/?phone="+ vendorPhoneNumber + "&text=" + url.QueryEscape(message)
	whatsappAPI += "&type=phone_number&app_absent=0"
	
	c.JSON(200, gin.H{
		"status": "success",
		"result": map[string]interface{}{
			"order": order,
			"details": orderDetails,
			"messageLink": whatsappAPI,
		},
		"errors": nil,
		"description": "Berhasil menyusun notifikasi untuk vendor untuk dikirim via Whatsapp.",
	})
}

func sanitizePhoneNumber(number string) (string, error) {
	var phoneNumber string = ""
	var errNumberNotValid error = errors.New("nomor telepon tidak valid: nomor telepon tidak mengikuti standar nomor telepon 08xx/628xx")
	var nonAlphanumericRegex = regexp.MustCompile(`\D+`)
	phoneNumber += nonAlphanumericRegex.ReplaceAllString(number, "")

	if len(phoneNumber) > 13 {
		errNumberTooLong := errors.New("nomor telepon terlalu panjang: ada kemungkinan merupakan gabungan dari banyak nomor")
		return "", errNumberTooLong
	}

	if phoneNumber[0:2] == "62" {
		return phoneNumber, nil
	}

	if phoneNumber[0:2] == "08" {
		prefix := "62"
		phoneNumber = prefix + strings.TrimPrefix(phoneNumber, "0")

		return phoneNumber, nil
	}

	return "", errNumberNotValid
}