package main

import (
	"testing"
	"net/http"
	"io"
	"net/http/httptest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/adeindriawan/itsfood-administration/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	return r
}

func TestDashboard(t *testing.T) {
	mockResponse := "Hello from dashboard"
	r := SetupRouter()
	r.GET("/admin", controllers.Dashboard)
	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	responseData, _ := io.ReadAll(w.Body)
	assert.Equal(t, mockResponse, string(responseData))
	assert.Equal(t, http.StatusOK, w.Code)
}