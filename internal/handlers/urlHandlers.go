package handlers

import (
	"errors"
	"net/http"
	"pranayteja31/Urlshortener/internal/services"
	"strconv"

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

func (h *URLHandler) CreateShortUrl(c *gin.Context) {
	orgUrl := c.PostForm("orgUrl")
	if orgUrl == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message":"The url field is empty", "error": "or_url is required"})
		return
	}
	expStr := c.PostForm("exp")
	exp,err := strconv.Atoi(expStr)
	if err != nil {
        exp = 30 // Default expiry if left blank or invalid
    }

	createdUrl, err := h.service.CreateShortURL(orgUrl,exp)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message":"Failed to create URL", "error":err.Error()})
		return
	}
	c.JSON(http.StatusCreated,gin.H{"message": "URL created Successfully!", "data": createdUrl})

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
