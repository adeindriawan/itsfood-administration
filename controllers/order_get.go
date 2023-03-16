package controllers

import (
	"strconv"
	"strings"
	"time"

	"github.com/adeindriawan/itsfood-administration/models"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/gin-gonic/gin"
)

func GetOrders(c *gin.Context) {
	var orders []OrderResult
	var messages = []string{}

	params := c.Request.URL.Query()
	idsParam, doesIdsParamExist := params["ids"]
	lengthParam, doesLengthParamExist := params["length"]
	pageParam, doesPageParamExist := params["page"]
	statusParam, doesStatusParamExist := params["status"]
	startOrderDateParam, doesStartOrderDateParamExist := params["order_date[start]"]
	endOrderDateParam, doesEndOrderDateParamExist := params["order_date[end]"]
	startDeliveryDateParam, doesStartDeliveryDateParamExist := params["delivery_date[start]"]
	endDeliveryDateParam, doesEndDeliveryDateParamExist := params["delivery_date[end]"]
	purposeParam, doesPurposeParamExist := params["purpose"]
	customerParam, doesCustomerParamExist := params["customer"]
	unitParam, doesUnitParamExist := params["unit"]
	paymentFromCustomerParam, doesPaymentFromCustomerParamExist := params["payment_from_customer"]
	paymentToVendorParam, doesPaymentToVendorParamExist := params["payment_to_vendor"]

	orderQuery := services.DB.Debug().Table("orders").
		Joins("LEFT JOIN order_details ON order_details.order_id = orders.id").
		Joins("LEFT JOIN customers ON customers.id = orders.ordered_by").
		Joins("LEFT JOIN users ON users.id = customers.user_id").
		Joins("LEFT JOIN units ON units.id = customers.unit_id").
		Select(`
			orders.id AS ID,
			orders.ordered_for AS OrderedFor,
			orders.ordered_to AS OrderedTo,
			orders.purpose AS Purpose,
			orders.status AS Status,
			orders.num_of_menus AS NumOfMenus,
			orders.qty_of_menus AS QtyOfMenus,
			users.name AS CustomerName,
			users.phone AS CustomerPhone,
			units.name AS CustomerUnit,
			orders.created_at AS CreatedAt,
			COUNT(order_details.paid_to_vendor_at) AS TotalMenuPaidToVendor,
			COUNT(order_details.id) AS TotalMenu
		`)

	if doesStatusParamExist {
		status := statusParam[0]
		orderQuery = orderQuery.Where("orders.status IN ?", strings.Split(status, ","))
	}

	if doesStartOrderDateParamExist {
		startOrderDate := startOrderDateParam[0]
		orderQuery = orderQuery.Where("DATE(orders.created_at) >= ?", startOrderDate)
	}

	if doesEndOrderDateParamExist {
		endOrderDate := endOrderDateParam[0]
		orderQuery = orderQuery.Where("DATE(orders.created_at) <= ?", endOrderDate)
	}

	if doesStartDeliveryDateParamExist {
		startDeliveryDate := startDeliveryDateParam[0]
		orderQuery = orderQuery.Where("DATE(orders.ordered_for) >= ?", startDeliveryDate)
	}

	if doesEndDeliveryDateParamExist {
		endDeliveryDate := endDeliveryDateParam[0]
		orderQuery = orderQuery.Where("DATE(orders.ordered_for) <= ?", endDeliveryDate)
	}

	if doesPurposeParamExist {
		purpose := purposeParam[0]
		orderQuery = orderQuery.Where("orders.purpose LIKE ?", "%"+purpose+"%")
	}

	if doesIdsParamExist {
		ids := idsParam[0]
		orderQuery = orderQuery.Where("orders.id IN ?", strings.Split(ids, ","))
	}

	if doesPaymentFromCustomerParamExist {
		paidFromCustomer := paymentFromCustomerParam[0]
		if paidFromCustomer == "paid" {
			orderQuery = orderQuery.Where("orders.paid_by_customer_at IS NOT NULL")
		}

		if paidFromCustomer == "unpaid" {
			orderQuery = orderQuery.Where("orders.paid_by_customer_at IS NULL")
		}
	}

	if doesPaymentToVendorParamExist {
		paymentToVendor := paymentToVendorParam[0]

		orderQuery = orderQuery.Where("orders.status != 'Cancelled'").Where("order_details.status != 'Cancelled'")

		if paymentToVendor == "unpaid" {
			orderQuery = orderQuery.Having("TotalMenuPaidToVendor = 0")
		}

		if paymentToVendor == "partially-paid" {
			orderQuery = orderQuery.Having("TotalMenuPaidToVendor <> TotalMenu AND TotalMenuPaidToVendor <> 0")
		}

		if paymentToVendor == "paid" {
			orderQuery = orderQuery.Having("TotalMenuPaidToVendor = TotalMenu")
		}
	}

	if doesCustomerParamExist {
		customer, _ := strconv.Atoi(customerParam[0])
		orderQuery = orderQuery.Where("orders.ordered_by", customer)
	}

	if doesUnitParamExist {
		unit, _ := strconv.Atoi(unitParam[0])
		orderQuery = orderQuery.Where("customers.unit_id", unit)
	}

	orderQuery.Group("orders.id")
	var totalRows int64
	orderQuery.Count(&totalRows)

	if doesLengthParamExist {
		length, err := strconv.Atoi(lengthParam[0])
		if err != nil {
			messages = append(messages, "Parameter Length tidak dapat dikonversi ke integer")
		} else {
			orderQuery = orderQuery.Limit(length)
		}
	}

	if doesPageParamExist {
		if doesLengthParamExist {
			page, _ := strconv.Atoi(pageParam[0])
			length, _ := strconv.Atoi(lengthParam[0])
			offset := (page - 1) * length
			orderQuery = orderQuery.Offset(offset)
		} else {
			messages = append(messages, "Tidak ada parameter Length, maka parameter Page diabaikan.")
		}
	}

	orderQuery.Scan(&orders)
	rowsCount := orderQuery.RowsAffected

	if orderQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      orderQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query.",
		})
		return
	}

	orderData := map[string]interface{}{
		"data":       orders,
		"rows_count": rowsCount,
		"total_rows": totalRows,
	}

	c.JSON(200, gin.H{
		"status":      "success",
		"result":      orderData,
		"errors":      messages,
		"description": "Berhasil mengambil data order",
	})
}

func GetOrder(c *gin.Context) {

	type ExtraCost struct {
		Amount uint64 `json:"amount"`
		Reason string `json:"reason"`
		Issuer string `json:"issuer"`
	}

	type Discount struct {
		Amount uint64 `json:"amount"`
		Reason string `json:"reason"`
		Issuer string `json:"issuer"`
	}

	type OrderDetail struct {
		ID         uint64      `json:"id"`
		Qty        uint        `json:"qty"`
		Price      uint64      `json:"price"`
		COGS       uint64      `json:"cogs"`
		Note       string      `json:"note"`
		Status     string      `json:"status"`
		CreatedAt  time.Time   `json:"created_at"`
		UpdatedAt  time.Time   `json:"updated_at"`
		CreatedBy  string      `json:"created_by"`
		MenuId     uint64      `json:"menu_id"`
		MenuName   string      `json:"menu_name"`
		VendorName string      `json:"vendor_name"`
		ExtraCosts []ExtraCost `json:"extra_costs"`
		Discounts  []Discount  `json:"discounts"`
	}

	type OrderInformationResult struct {
		ID             uint64        `json:"id"`
		OrderedFor     time.Time     `json:"ordered_for"`
		OrderedTo      string        `json:"ordered_to"`
		Purpose        string        `json:"purpose"`
		Activity       string        `json:"activity"`
		SourceOfFund   string        `json:"source_of_fund"`
		PaymentOption  string        `json:"payment_option"`
		Info           string        `json:"info"`
		Status         string        `json:"status"`
		CreatedAt      time.Time     `json:"created_at"`
		UpdatedAt      time.Time     `json:"updated_at"`
		CreatedBy      string        `json:"created_by"`
		PurchaseAmount int64         `json:"purchase_amount"`
		SalesAmount    int64         `json:"sales_amount"`
		CustomerId     uint64        `json:"customer_id"`
		CustomerName   string        `json:"customer_name"`
		CustomerUnit   string        `json:"customer_unit"`
		CustomerPhone  string        `json:"customer_phone"`
		CustomerEmail  string        `json:"customer_email"`
		OrderDetails   []OrderDetail `json:"order_details"`
	}

	var orderInformation OrderInformationResult
	var order models.Order
	var customer models.Customer
	var orderDetailsRaw []models.OrderDetail

	var messages = []string{}

	orderId, notValidId := strconv.Atoi(c.Param("id"))

	if notValidId != nil {
		c.JSON(400, gin.H{
			"status":      "failed",
			"message":     "ID tidak valid",
			"description": "Gagal mengambil data order",
		})
		return
	}

	orderQuery := services.DB.First(&order, orderId)

	if orderQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      orderQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query order.",
		})
		return
	}

	customerQuery := services.DB.Preload("User").Preload("Unit").First(&customer, order.OrderedBy)

	if customerQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      customerQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query customer.",
		})
		return
	}

	orderDetailsQuery := services.DB.Preload("Menu.Vendor.User").
		Preload("Discounts").
		Preload("Costs").
		Find(&orderDetailsRaw, "order_id = ?", order.ID)

	if orderDetailsQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      orderDetailsQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query order details.",
		})
		return
	}

	var purchaseAmount int64
	var salesAmount int64

	var orderDetails []OrderDetail

	for _, od := range orderDetailsRaw {

		var extraCosts []ExtraCost
		var discounts []Discount

		var orderDetail OrderDetail
		var totalExtraCost int64
		var totalVendorExtraCost int64
		var totalDiscount int64
		var totalVendorDiscount int64

		for _, cost := range od.Costs {
			if cost.Issuer == "Vendor" {
				totalVendorExtraCost += int64(cost.Amount)
			}

			totalExtraCost += int64(cost.Amount)
			extraCosts = append(extraCosts, ExtraCost{
				Amount: uint64(cost.Amount),
				Reason: cost.Reason,
				Issuer: cost.Issuer,
			})
		}

		for _, discount := range od.Discounts {
			if discount.Issuer == "Vendor" {
				totalVendorDiscount += int64(discount.Amount)
			} else {
				totalDiscount += int64(discount.Amount)
			}
			discounts = append(discounts, Discount{
				Amount: uint64(discount.Amount),
				Reason: discount.Reason,
				Issuer: discount.Issuer,
			})
		}

		orderDetail.ID = od.ID
		orderDetail.Qty = od.Qty
		orderDetail.Price = od.Price
		orderDetail.COGS = od.Price
		orderDetail.Note = od.Note
		orderDetail.Status = od.Status
		orderDetail.CreatedAt = od.CreatedAt
		orderDetail.UpdatedAt = od.UpdatedAt
		orderDetail.CreatedBy = od.CreatedBy
		orderDetail.MenuId = od.MenuID
		orderDetail.MenuName = od.Menu.Name
		orderDetail.VendorName = od.Menu.Vendor.User.Name
		orderDetail.ExtraCosts = extraCosts
		orderDetail.Discounts = discounts

		orderDetails = append(orderDetails, orderDetail)

		purchaseAmount += (int64(od.COGS) * int64(od.Qty)) + totalVendorExtraCost - totalVendorDiscount
		salesAmount += (int64(od.Price) * int64(od.Qty)) + totalExtraCost - totalDiscount

	}

	orderInformation.ID = order.ID
	orderInformation.OrderedFor = order.OrderedFor
	orderInformation.OrderedTo = order.OrderedTo
	orderInformation.Purpose = order.Purpose
	orderInformation.Activity = order.Activity
	orderInformation.SourceOfFund = order.SourceOfFund
	orderInformation.PaymentOption = order.PaymentOption
	orderInformation.Info = order.Info
	orderInformation.Status = order.Status
	orderInformation.CreatedAt = order.CreatedAt
	orderInformation.UpdatedAt = order.UpdatedAt
	orderInformation.CreatedBy = order.CreatedBy
	orderInformation.PurchaseAmount = purchaseAmount
	orderInformation.SalesAmount = salesAmount
	orderInformation.CustomerId = customer.ID
	orderInformation.CustomerName = customer.User.Name
	orderInformation.CustomerUnit = customer.Unit.Name
	orderInformation.CustomerPhone = customer.User.Phone
	orderInformation.CustomerEmail = customer.User.Email
	orderInformation.OrderDetails = orderDetails

	orderInformationData := map[string]interface{}{
		"data": orderInformation,
	}

	c.JSON(200, gin.H{
		"status":      "success",
		"result":      orderInformationData,
		"errors":      messages,
		"description": "Berhasil mengambil data order",
	})
}
