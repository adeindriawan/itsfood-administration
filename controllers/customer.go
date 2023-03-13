package controllers

import (
	"strconv"

	"github.com/adeindriawan/itsfood-administration/models"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/gin-gonic/gin"
)

type CustomerResult struct {
	models.Customer
	Name     string `json:"name"`
	UnitName string `json:"unit_name"`
}

func GetCustomers(c *gin.Context) {
	var customers []CustomerResult
	var messages = []string{}

	params := c.Request.URL.Query()

	lengthParam, doesLengthParamExist := params["length"]
	pageParam, doesPageParamExist := params["page"]
	searchParam, doesSearchParamExist := params["search"]

	customerQuery := services.DB.Debug().Table("customers").
		Joins("LEFT JOIN users ON users.id = customers.user_id").
		Joins("LEFT JOIN units ON units.id = customers.unit_id").
		Select(`
			customers.id AS ID,
			customers.user_id AS UserID,
			users.name AS Name,
			customers.type AS Type,
			customers.unit_id AS UnitID,
			units.name AS UnitName,
			customers.status AS Status,
			customers.created_by AS CreatedBy,
			customers.created_at AS CreatedAt,
			customers.updated_at AS UpdatedAt
		`)

	if doesSearchParamExist {
		search := searchParam[0]
		customerQuery = customerQuery.Where("users.name LIKE ?", "%"+search+"%")
	}

	var totalRows int64
	customerQuery.Count(&totalRows)

	if doesLengthParamExist {
		length, err := strconv.Atoi(lengthParam[0])
		if err != nil {
			messages = append(messages, "Parameter Length tidak dapat dikonversi ke integer")
		} else {
			customerQuery = customerQuery.Limit(length)
		}
	}

	if doesPageParamExist {
		if doesLengthParamExist {
			page, _ := strconv.Atoi(pageParam[0])
			length, _ := strconv.Atoi(lengthParam[0])
			offset := (page - 1) * length
			customerQuery = customerQuery.Offset(offset)
		} else {
			messages = append(messages, "Tidak ada parameter Length, maka parameter Page diabaikan.")
		}
	}

	customerQuery.Scan(&customers)
	rowsCount := customerQuery.RowsAffected

	if customerQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      customerQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query.",
		})
	}

	customerData := map[string]interface{}{
		"data":       customers,
		"rows_count": rowsCount,
		"total_rows": totalRows,
	}

	c.JSON(200, gin.H{
		"status":      "success",
		"result":      customerData,
		"errors":      messages,
		"description": "Berhasil mengambil data customer.",
	})
}
