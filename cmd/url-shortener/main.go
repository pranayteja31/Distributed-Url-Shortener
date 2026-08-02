package main

//this is the main entry point of the application
import (
	"pranayteja31/Urlshortener/internal/config"

	"github.com/gin-gonic/gin"
)



func main() {
	cfg := config.MustLoad()
	
	//init the router
	router := gin.Default()
	router.GET("/",func(ctx *gin.Context) {
		ctx.JSON(200,gin.H{"message":"hello"})
	})

	router.Run(cfg.HTTPServer.Addr)
	
}
