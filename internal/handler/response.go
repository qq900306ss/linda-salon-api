package handler

import "github.com/gin-gonic/gin"

// errorBody is the error payload inside the response envelope.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a success envelope with the given HTTP status.
func OK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

// Fail writes an error envelope with the given HTTP status.
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"success": false, "error": errorBody{Code: code, Message: message}})
}
