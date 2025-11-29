package main

import (
	"log"

	"task_manager/router"
)

func main() {
	r := router.Setup()
	log.Println("🚀 Task Manager API running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
