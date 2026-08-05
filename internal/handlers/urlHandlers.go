package handlers

import (
	"net/http"
	"pranayteja31/Urlshortener/internal/models"
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

func (h *URLHandler)Redirect(c *gin.Context) {
	shortCode := c.Query("shortCode")
	orgCode,err := h.service.RedirectURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"message": "Unable to Fetch Details","error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect,orgCode)

}

func (h *URLHandler)GetUrl(c *gin.Context) {
	idStr := c.Query("id")
	id,err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Invalid Id","error": err.Error()})
		return
	}
	urlDetails, err := h.service.GetURL(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"message": "URL Not Found","error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message": "ID found", "data": urlDetails})

}

func (h *URLHandler)ListUrls(c *gin.Context) {
	UrlList, err := h.service.ListURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"message": "Something went Wrong", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message":"All Urls fetched successfully", "data": UrlList})
}

func (h *URLHandler)UpdateUrl(c *gin.Context) {
	idStr := c.Query("id")
	orgUrl := c.PostForm("orgUrl")
	expStr := c.PostForm("exp")

	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Invalid Id", "error": err.Error()})
		return
	}
	exp,err := strconv.Atoi(expStr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Invalid Expiry", "error": err.Error()})
	}

	newUrl := models.URL{
		ID: id,
		OriginalURL: orgUrl,
	}
	updateUrl, err := h.service.UpdateURL(&newUrl,exp)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Update Unsuccessful", "error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted,gin.H{"message": "Update Successful", "data": updateUrl})
}
func DeletUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message":"deletes the short url"})
}
