package handler

import (
	"net/http"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/application/service"
	"github.com/dmitokk/FragmentsBE/internal/http/middleware"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
	minioClient *minio.Client
}

func NewUserHandler(userService *service.UserService, minioClient *minio.Client) *UserHandler {
	return &UserHandler{userService: userService, minioClient: minioClient}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	resp, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := c.PostForm("name")

	var avatarURL string
	file, header, err := c.Request.FormFile("avatar")
	if err == nil {
		defer file.Close()

		objectName, uploadErr := h.minioClient.UploadAvatar(c.Request.Context(), userID.String(), header.Filename, file, header.Size)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
			return
		}

		avatarURL = "/api/files/" + objectName
	}

	resp, err := h.userService.UpdateProfile(c.Request.Context(), userID, &dto.UpdateProfileRequest{
		Name:      name,
		AvatarURL: avatarURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
