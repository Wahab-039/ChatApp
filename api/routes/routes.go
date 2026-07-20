package routes

import (
	"github.com/Wahab-039/ChatApp/api/handlers"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/health", handlers.Health)
}
