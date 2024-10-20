package api

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"
	"github.com/skyline93/mediabox/internal/fs"
	"github.com/skyline93/mediabox/internal/mediabox"

	"github.com/gin-gonic/gin"
)

// UploadPhoto godoc
//
//	@Summary		Upload photo
//	@Description	upload photo
//	@Tags			Photos
//	@Accept			json
//	@Produce		json
//	@Router			/api/v1/photo/upload [post]
//	@Param			album_id	formData	int		true	"album id"
//	@Param			file		formData	file	true	"the file to upload"
//	@Success		200			{object}	Response
func UploadPhoto(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		user := entity.FindUser(username.(string))

		albumIDStr := c.PostForm("album_id")
		if albumIDStr == "" {
			c.JSON(http.StatusBadRequest, Error(400, "Album ID is required"))
			return
		}
		albumID, err := strconv.Atoi(albumIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid album ID"))
			return
		}

		var album *entity.Album
		for _, alb := range user.Albums {
			if alb.ID == uint(albumID) {
				album = &alb
				break
			}
		}

		if album == nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid album ID"))
			return
		}

		files := c.Request.MultipartForm.File["files[]"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, Error(400, "No files provided"))
			return
		}

		for _, file := range files {
			uniqueFileName := mediabox.GenerateUniqueFilename(file.Filename)

			photo := &entity.Photo{
				Name:     filepath.Base(file.Filename),
				FileName: filepath.Base(uniqueFileName),
				FileSize: file.Size,
				FileType: strings.Split(file.Header.Get("Content-Type"), ";")[0],
				AlbumID:  album.ID,
				UserID:   user.ID,
				Ext:      fs.FileType(strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))),
			}

			src, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, Error(400, "open src file failed"))
				return
			}
			defer src.Close()

			if err := mediabox.UploadPhoto(user.UUID, album.UUID, uniqueFileName, src, conf); err != nil {
				c.JSON(http.StatusInternalServerError, Error(400, "Failed to save file"))
				return
			}

			_, err = photo.Create(uint(albumID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, Error(400, "Failed to create photo"))
				return
			}
		}

		c.JSON(http.StatusOK, Success("Photos uploaded successfully"))
	}

	router.POST("/photo/upload", handler)
}

// ImportPhoto godoc
//
//	@Summary		Import photo
//	@Description	import photo
//	@Tags			Photos
//	@Accept			json
//	@Produce		json
//	@Router			/api/v1/photo/import [post]
//	@Success		200	{object}	Response
func ImportPhoto(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		if err := mediabox.ImportOriginals(username.(string), conf); err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Failed to import photo"))
			return
		}

		if err := mediabox.ImportOriginalsFromWebDAV(username.(string), conf); err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Failed to import photo"))
			return
		}

		c.JSON(http.StatusOK, Success("import successfully"))
	}

	router.POST("/photo/import", handler)
}

type ListPhotosResponse struct {
	Items     []entity.Photo `json:"items"`
	TotalNum  int            `json:"total_num"`
	TotalPage int            `json:"total_page"`
	Page      int            `json:"page"`
	Limit     int            `json:"limit"`
}

// GetPhotos godoc
//
//	@Summary		Get photos
//	@Description	get photos
//	@Tags			Photos
//	@Accept			json
//	@Produce		json
//	@Router			/api/v1/photo [get]
//	@Param			album_id	query		int	true	"album id"
//	@Success		200			{object}	Response
func ListPhotos(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		user := entity.FindUser(username.(string))

		albumID, _ := strconv.Atoi(c.Query("album_id"))
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))

		var album *entity.Album
		for _, alb := range user.Albums {
			if alb.ID == uint(albumID) {
				album = &alb
				break
			}
		}

		if album == nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid album ID"))
			return
		}

		photos, totalNum, totalPage, err := entity.ListPhotos(username.(string), album.ID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Failed to import photo"))
			return
		}

		for i := range photos {
			link := router.BasePath() + fmt.Sprintf("/photo/thumbnail/%s/%s.jpg", filepath.Join(user.UUID, album.UUID), photos[i].FileName)
			photos[i].Link = link
		}

		c.JSON(http.StatusOK, ListPhotosResponse{
			Items:     photos,
			TotalNum:  totalNum,
			TotalPage: totalPage,
			Page:      page,
			Limit:     limit,
		})
	}

	router.GET("/photo", handler)
}

func GetPhotoFile(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		path := c.Param("path")

		file, err := os.ReadFile(filepath.Join(conf.StoragePath, "thumbnails", path))
		if err != nil {
			if errors.Is(err, iofs.ErrNotExist) {
				c.JSON(http.StatusNotFound, Error(400, "File not found"))
				return
			}
			c.JSON(http.StatusInternalServerError, Error(400, "Internal server error"))
			return
		}

		contentType := http.DetectContentType(file)

		c.Data(http.StatusOK, contentType, file)
	}

	router.GET("/photo/thumbnail/*path", handler)
}

func DeletePhotos(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		var ids []uint
		if err := c.BindJSON(&ids); err != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid request payload"))
			return
		}

		err := entity.DeletePhotos(ids)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Failed to delete photos"))
			return
		}

		c.JSON(http.StatusOK, Success("Photos marked as deleted"))
	}

	router.DELETE("/photo", handler)
}
