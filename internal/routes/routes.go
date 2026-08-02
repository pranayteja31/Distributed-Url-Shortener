package routes

import (
	"pranayteja31/Urlshortener/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine){
	//routing the routes with handlers

	//creation of the short url
	router.POST("/create-url",handlers.CreateShortUrl)

	//creation of the short url
	router.GET("/redirect",handlers.Redirect)

	//creation of the short url
	router.POST("/get-url",handlers.GetUrl)

	//creation of the short url
	router.GET("/list-url",handlers.ListUrls)

	//creation of the short url
	router.PUT("/update-url",handlers.UpdateUrl)

	//creation of the short url
	router.DELETE("/delete-url",handlers.DeletUrl)
}