package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUrlAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"creates the short url"})
}

func GetGlobalAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"redirects the short url"})
}

func GetTopUrls(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"gives the actual url of the short url"})
}

