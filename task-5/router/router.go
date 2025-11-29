package router

import (
	"github.com/gin-gonic/gin"
	"task_manager/controllers"
	"task_manager/data"
)

func Setup() *gin.Engine {
	r := gin.Default()

	// Simple CORS for local testing; adjust as needed
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/")

	taskService := data.NewInMemoryTaskService()
	taskController := controllers.NewTaskController(taskService)
	taskController.Register(api)

	// Health endpoint (handy in Postman)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
