package response

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) { c.JSON(200, gin.H{"success": true, "data": data}) }
func Created(c *gin.Context, data any) { c.JSON(201, gin.H{"success": true, "data": data}) }
func Err(c *gin.Context, code int, msg string) { c.JSON(code, gin.H{"success": false, "error": msg}) }
