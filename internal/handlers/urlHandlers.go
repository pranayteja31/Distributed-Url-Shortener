package handlers

import (
	"net/http"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

//struct
type URLHandler struct {
	service *services.URLServices
}
//constructor
func NewHandler(service *services.URLServices) *URLHandler {
	return &URLHandler{
		service: service,
	}
}
//creation
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
//redirect
func (h *URLHandler)Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	orgCode,err := h.service.RedirectURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"message": "Unable to Fetch Details","error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect,orgCode)

}
//get url by id
func (h *URLHandler)GetUrl(c *gin.Context) {
	idStr := c.Param("id")
	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Invalid Id","error": err.Error()})
		return
	}
	urlDetails, err := h.service.GetURL(id)
	if err != nil {
		c.JSON(http.StatusNotFound,gin.H{"message": "URL Not Found","error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message": "ID found", "data": urlDetails})

}
//list all available urls
func (h *URLHandler)ListUrls(c *gin.Context) {
	UrlList, err := h.service.ListURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"message": "Something went Wrong", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message":"All Urls fetched successfully", "data": UrlList})
}
//update url details
func (h *URLHandler)UpdateUrl(c *gin.Context) {
	idStr := c.Param("id")
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

//delete url
func (h *URLHandler)DeleteUrl(c *gin.Context) {
	idStr := c.Param("id")
	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"message": "Invalid Id", "error": err.Error()})
	}
	if err = h.service.DeleteURL(id); err != nil {
		c.JSON(http.StatusNotFound,gin.H{"message": "Delete Unsuccessful", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message": "URL Deleted Successful"})

}
