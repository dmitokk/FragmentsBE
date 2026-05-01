package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/application/service"
	"github.com/dmitokk/FragmentsBE/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FragmentHandler struct {
	fragmentService *service.FragmentService
}

func NewFragmentHandler(fragmentService *service.FragmentService) *FragmentHandler {
	return &FragmentHandler{fragmentService: fragmentService}
}

func (h *FragmentHandler) Create(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateFragmentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	var photos []io.Reader
	var photoSizes []int64
	if photoFiles, ok := form.File["photos"]; ok {
		for _, photoFile := range photoFiles {
			file, err := photoFile.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open photo file"})
				return
			}
			defer file.Close()
			photos = append(photos, file)
			photoSizes = append(photoSizes, photoFile.Size)
		}
	}

	var sound io.Reader
	var soundSize int64
	if soundFiles, ok := form.File["sound"]; ok && len(soundFiles) > 0 {
		soundFile := soundFiles[0]
		file, err := soundFile.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to open sound file"})
			return
		}
		defer file.Close()
		sound = file
		soundSize = soundFile.Size
	}

	resp, err := h.fragmentService.Create(c.Request.Context(), userID, &req, photos, photoSizes, sound, soundSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *FragmentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fragment ID"})
		return
	}

	resp, err := h.fragmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FragmentHandler) List(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	latStr := c.DefaultQuery("lat", "0")
	lngStr := c.DefaultQuery("lng", "0")
	radiusStr := c.DefaultQuery("radius", "0")

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)
	radius, _ := strconv.ParseFloat(radiusStr, 64)

	resp, err := h.fragmentService.List(c.Request.Context(), userID, lat, lng, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FragmentHandler) Update(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fragment ID"})
		return
	}

	var req dto.UpdateFragmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fragment, err := h.fragmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fragment not found"})
		return
	}

	if fragment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	resp, err := h.fragmentService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *FragmentHandler) Delete(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fragment ID"})
		return
	}

	fragment, err := h.fragmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fragment not found"})
		return
	}

	if fragment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	err = h.fragmentService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}