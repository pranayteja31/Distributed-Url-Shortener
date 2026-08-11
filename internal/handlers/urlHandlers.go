package handlers

import (
	"errors"
	"net/http"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/services"
	"pranayteja31/Urlshortener/internal/utils"
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
		c.JSON(http.StatusBadRequest, gin.H{"message":"The url field is empty", "error": "orgUrl is required"})
		return
	}
	expStr := c.PostForm("exp")
	exp := 30
	if expStr != "" {
        exp,err := strconv.Atoi(expStr)
		if err != nil || exp <= 0 {
			c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid expiry"})
			return
		}
		
    }

	createdUrl, err := h.service.CreateShortURL(orgUrl,exp)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrEmptyURL):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL cannot be empty",
			})
			return

		case errors.Is(err, utils.ErrURLTooLong):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL exceeds maximum length",
			})
			return

		case errors.Is(err, utils.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid URL",
			})
			return

		case errors.Is(err, utils.ErrUnsupportedScheme):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "only HTTP and HTTPS URLs are supported",
			})
			return

		case errors.Is(err, utils.ErrMissingURLHost):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL host is missing",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create URL",
			})
			return
		}
	}
	c.JSON(http.StatusCreated,gin.H{"message": "URL created Successfully!", "data": createdUrl})

}
//redirect
func (h *URLHandler)Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest,gin.H{
			"error": "Shortcode required",
		})
		return
	}
	OriginalURL,err := h.service.RedirectURL(shortCode,c.ClientIP(),c.GetHeader("User_Agent"),c.GetHeader("Referer"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return

		case errors.Is(err, services.ErrURLNotFound):
			c.JSON(http.StatusGone, gin.H{
				"error": "URL has expired",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
	}

	c.Redirect(http.StatusTemporaryRedirect,OriginalURL)

}
//get url by id
func (h *URLHandler)GetUrl(c *gin.Context) {
	idStr := c.Param("id")
	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest,gin.H{"error": "Invalid ID"})
		return
	}
	urlDetails, err := h.service.GetURL(id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}
	}
	c.JSON(http.StatusOK,gin.H{"message": "URL found", "data": urlDetails})

}
//list all available urls
func (h *URLHandler)ListUrls(c *gin.Context) {
	UrlList, err := h.service.ListURLs()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"message": "Something went Wrong", "error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message":"All URLs fetched successfully", "data": UrlList})
}
//update url details
func (h *URLHandler)UpdateUrl(c *gin.Context) {
	idStr := c.Param("id")
	orgUrl := c.PostForm("orgUrl")
	expStr := c.PostForm("exp")

	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil || id <=0 {
		c.JSON(http.StatusBadRequest,gin.H{"error": "Invalid ID"})
		return
	}
	if orgUrl == "" {
		c.JSON(http.StatusBadRequest,gin.H{"error": "Original URL required"})
		return
	}

	exp,err := strconv.Atoi(expStr)
	if err != nil || exp <= 0 {
		c.JSON(http.StatusBadRequest,gin.H{"error": "Invalid Expiry"})
		return
	}

	newUrl := models.URL{
		ID: id,
		OriginalURL: orgUrl,
	}
	updateUrl, err := h.service.UpdateURL(&newUrl,exp)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrEmptyURL):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL cannot be empty",
			})
			return

		case errors.Is(err, utils.ErrURLTooLong):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL exceeds maximum length",
			})
			return

		case errors.Is(err, utils.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid URL",
			})
			return

		case errors.Is(err, utils.ErrUnsupportedScheme):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "only HTTP and HTTPS URLs are supported",
			})
			return

		case errors.Is(err, utils.ErrMissingURLHost):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL host is missing",
			})
			return
		case errors.Is(err, services.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update URL",
			})
			return
		}
	}
	c.JSON(http.StatusAccepted,gin.H{"message": "Update Successful", "data": updateUrl})
}

//delete url
func (h *URLHandler)DeleteUrl(c *gin.Context) {
	idStr := c.Param("id")
	id,err := strconv.ParseInt(idStr,10,64)
	if err != nil || id<=0 {
		c.JSON(http.StatusBadRequest,gin.H{"error": "Invalid ID"})
		return
	}
	if err = h.service.DeleteURL(id); err != nil {
		switch {
		case errors.Is(err, services.ErrURLNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "URL not found",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete URL",
			})
			return
		}
	}
	c.JSON(http.StatusOK,gin.H{"message": "URL Deleted Successful"})

}
