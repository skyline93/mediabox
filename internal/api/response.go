package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "OK",
		Data:    data,
	}
}

func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			var err error
			switch c.Errors[0].Err.(type) {
			case *gin.Error:
				err = c.Errors[0].Err.(*gin.Error).Err
			default:
				err = c.Errors[0].Err
			}

			code := http.StatusInternalServerError
			switch c.Writer.Status() {
			case http.StatusNotFound:
				code = http.StatusNotFound
			case http.StatusBadRequest:
				code = http.StatusBadRequest
			}

			c.AbortWithStatusJSON(code, Error(code, err.Error()))
		}
	}
}
