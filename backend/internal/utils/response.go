package utils

import (
	"net/http"
	"schedule-system-v2/backend/internal/model"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, model.Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func ErrorWithStatus(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, model.Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func PageSuccess(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, model.PageResponse{
		Code:    200,
		Message: "success",
		Data:    list,
		Total:   total,
	})
}
