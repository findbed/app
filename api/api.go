package api

import (
	"github.com/gin-gonic/gin"
)

func APIRouter(engine *gin.Engine) {
	engine.GET("/api/list", list)
}

func list(c *gin.Context) {
	c.JSON(200, gin.H{
		"data": []gin.H{
			{
				"id":      "13123",
				"country": "russia",
			},
			{
				"id":      "4233",
				"country": "finland",
			},
		},
	})
}
