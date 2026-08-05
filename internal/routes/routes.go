package routes

import (
	"pranayteja31/Urlshortener/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, handler *handlers.URLHandler){
	//routing the routes with handlers

	v1 := router.Group("/api/v1")
	url := v1.Group("/url")
	//"/api/v1/..."
	//creation of the short url
	url.POST("/create", handler.CreateShortUrl)
	//get all the details about the url
	url.GET("/get/:id",handler.GetUrl)
	//list all urls
	url.GET("/list",handler.ListUrls)
	//update the url
	url.PUT("/update/:id",handler.UpdateUrl)
	//delete the url
	url.DELETE("/delete/:id",handlers.DeletUrl)

	//creation of the short url
	router.GET("/:shortCode",handler.Redirect)
}