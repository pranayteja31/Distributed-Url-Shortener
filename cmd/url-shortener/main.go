package main

//this is the main entry point of the application
import (
	"pranayteja31/Urlshortener/internal/config"
	"pranayteja31/Urlshortener/internal/db"
	"pranayteja31/Urlshortener/internal/handlers"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/routes"
	"pranayteja31/Urlshortener/internal/services"

	"github.com/gin-gonic/gin"
)



func main() {
	cfg := config.MustLoad()

	dbConn := db.DB_connection()
	defer dbConn.Close()

	//init the router
	router := gin.Default()

	//register repos with db
	repo := repository.NewRepository(dbConn)
	//register services with repos
	services := services.NewURLServices(repo)
	//register handlers with the services
	handlers := handlers.NewHandler(services)

	//registering the routers with handlers
	routes.RegisterRoutes(router, handlers)

	//start the server
	router.Run(cfg.HTTPServer.Addr)
}
