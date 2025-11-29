package router

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"task_manager/controllers"
	"task_manager/data"
)

func Setup() *gin.Engine {
	r := gin.Default()

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

	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getenv("MONGO_DB", "task_manager")
	taskColName := getenv("MONGO_TASKS_COLLECTION", "tasks")
	userColName := getenv("MONGO_USERS_COLLECTION", "users")

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil { log.Fatalf("mongo client init: %v", err) }
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Connect(ctx); err != nil { log.Fatalf("mongo connect: %v", err) }
		if err := client.Ping(ctx, nil); err != nil { log.Fatalf("mongo ping: %v", err) }
	}

	db := client.Database(dbName)
	taskCol := db.Collection(taskColName)
	userCol := db.Collection(userColName)

	ensureUserIndexes(userCol)

	taskSvc := data.NewTaskService(taskCol)
	userSvc := data.NewUserService(userCol)
	ctrl := controllers.NewController(taskSvc, userSvc)

	ctrl.RegisterRoutes(r)

	return r
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ensureUserIndexes(col *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("warn: failed to create username unique index: %v", err)
	}
}
