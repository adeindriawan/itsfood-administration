package controllers

import (
	"strconv"

	"github.com/adeindriawan/itsfood-administration/models"
	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/gin-gonic/gin"
)

type UnitResult = models.Unit

func GetUnits(c *gin.Context) {
	var units []UnitResult
	var messages = []string{}

	params := c.Request.URL.Query()

	lengthParam, doesLengthParamExist := params["length"]
	pageParam, doesPageParamExist := params["page"]
	searchParam, doesSearchParamExist := params["search"]

	unitQuery := services.DB.Table("units")

	if doesSearchParamExist {
		search := searchParam[0]
		unitQuery = unitQuery.Where("name LIKE ?", "%"+search+"%")
	}

	var totalRows int64
	unitQuery.Count(&totalRows)

	if doesLengthParamExist {
		length, err := strconv.Atoi(lengthParam[0])
		if err != nil {
			messages = append(messages, "Parameter Length tidak dapat dikonversi ke integer")
		} else {
			unitQuery = unitQuery.Limit(length)
		}
	}

	if doesPageParamExist {
		if doesLengthParamExist {
			page, _ := strconv.Atoi(pageParam[0])
			length, _ := strconv.Atoi(lengthParam[0])
			offset := (page - 1) * length
			unitQuery = unitQuery.Offset(offset)
		} else {
			messages = append(messages, "Tidak ada parameter Length, maka parameter Page diabaikan.")
		}
	}

	unitQuery.Scan(&units)
	rowsCount := unitQuery.RowsAffected

	if unitQuery.Error != nil {
		c.JSON(512, gin.H{
			"status":      "failed",
			"errors":      unitQuery.Error.Error(),
			"result":      nil,
			"description": "Gagal mengeksekusi query.",
		})
	}

	unitData := map[string]interface{}{
		"data":       units,
		"rows_count": rowsCount,
		"total_rows": totalRows,
	}

	c.JSON(200, gin.H{
		"status":      "success",
		"result":      unitData,
		"errors":      messages,
		"description": "Berhasil mengambil data unit.",
	})
}
