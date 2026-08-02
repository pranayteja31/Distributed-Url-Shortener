package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateShortUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"creates the short url"})
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
