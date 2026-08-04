package handlers

import (
	"net/http"
	"pranayteja31/Urlshortener/internal/services"

	"github.com/gin-gonic/gin"
)
type URLHandler struct {
	service *services.URLServices
}

func NewHandler(service *services.URLServices) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

func CreateShortUrl(c *gin.Context) {
	
}

func Redirect(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"redirects the short url"})
}

func GetUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"gives the actual url of the short url"})
}

func ListUrls(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"lists all of the short url"})
}
func UpdateUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"updates the created short url"})
}
func DeletUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"deletes the short url"})
}
