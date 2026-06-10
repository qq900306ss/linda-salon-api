// Local development server entrypoint (plain HTTP on :4000).
package main

import (
	"context"
	"log"
	"os"

	"github.com/qq900306ss/linda-salon-api/internal/app"
)

func main() {
	router, err := app.Initialize(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
