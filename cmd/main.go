package main

import (
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
	"database/sql"
	"fmt"
	"os"
	 _ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {

	url := os.Getenv("TURSO_DATABASE_URL")

	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}
	defer db.Close()

	router := gin.Default()

	api := router.Group("/api")
	routes.RegisterRoutes(api)

	router.Run("0.0.0.0:8080")
	//router.Run("localhost:8080")
}
