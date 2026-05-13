package handler

import (
	"net/http"
	"strings"

	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	minioClient *minio.Client
}

func NewFileHandler(minioClient *minio.Client) *FileHandler {
	return &FileHandler{minioClient: minioClient}
}

func (h *FileHandler) ServeFile(c *gin.Context) {
	filepath := c.Param("filepath")
	filepath = strings.TrimPrefix(filepath, "/")
	if filepath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
		return
	}

	reader, size, contentType, err := h.minioClient.GetFile(c.Request.Context(), filepath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	extraHeaders := map[string]string{
		"Cache-Control": "private, max-age=3600",
	}
	c.DataFromReader(http.StatusOK, size, contentType, reader, extraHeaders)
}
