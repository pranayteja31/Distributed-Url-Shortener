package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Urlshorten(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"creates the short url"})
}

