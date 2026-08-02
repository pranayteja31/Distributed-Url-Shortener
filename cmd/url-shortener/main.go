package main

//this is the main entry point of the application
import (
	"net/http"
	"pranayteja31/Urlshortener/internal/config"
	"pranayteja31/Urlshortener/internal/db"
	"pranayteja31/Urlshortener/internal/routes"

	"github.com/gin-gonic/gin"
)



func main() {
	cfg := config.MustLoad()
	//dbconnection
	db.DB_connection()
	defer db.DB.Close()
	
	//init the router
	router := gin.Default()
	//registering the routers created
	routes.RegisterRoutes(router)
	router.GET("/status", func(c *gin.Context) {
		var currentTime string
		
		// Query the database directly
		err := db.DB.QueryRow("SELECT NOW()").Scan(&currentTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":        "Database connected successfully!",
			"database_time": currentTime,
		})
	})

	router.Run(cfg.HTTPServer.Addr)
	
	
}
