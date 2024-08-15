package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/skyline93/mediabox/internal/config"

	"github.com/gin-gonic/gin"
)

func UploadFile(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, Error(400, "No file is received"))
			return
		}

		userDir := filepath.Join(conf.StoragePath, "uploads", username.(string))
		if _, err := os.Stat(userDir); os.IsNotExist(err) {
			err := os.MkdirAll(userDir, os.ModePerm)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Error(400, "Could not create user directory"))
				return
			}
		}

		filePath := filepath.Join(userDir, file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Could not save file"))
			return
		}

		c.JSON(http.StatusOK, Error(400, "File uploaded successfully"))
	}

	router.POST("/upload", handler)
}
