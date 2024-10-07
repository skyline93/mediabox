package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"

	"github.com/gin-gonic/gin"
)

type CreateAlbumRequest struct {
	AlbumName string `json:"album_name" binding:"required"`
}

// CreateAlbums godoc
//
//	@Summary		Create albums
//	@Description	create albums
//	@Tags			Albums
//	@Accept			json
//	@Produce		json
//	@Router			/api/v1/albums [post]
//	@Param			album	body		CreateAlbumRequest	true	"album info"
//	@Success		200		{object}	Response
func CreateAlbum(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		user := entity.FindUser(username.(string))
		if user == nil {
			c.JSON(http.StatusInternalServerError, Error(400, "user not found"))
			return
		}

		var json CreateAlbumRequest

		if c.Bind(&json) != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid request"))
			return
		}

		if entity.ExistsAlbum(json.AlbumName, user.Name) {
			c.JSON(http.StatusConflict, Error(400, "Album already exists"))
			return
		}

		id := uuid.New()
		album := entity.Album{Name: json.AlbumName, UUID: id.String()}
		alb, err := album.Create(json.AlbumName, user.ID)
		if err != nil {
			c.JSON(http.StatusConflict, Error(400, "Album create failed"))
			return
		}

		c.JSON(http.StatusOK, Success(fmt.Sprintf("album %d created successfully", alb.ID)))
	}

	router.POST("/albums", handler)
}

// GetAlbums godoc
//
//	@Summary		Get albums
//	@Description	get albums
//	@Tags			Albums
//	@Accept			json
//	@Produce		json
//	@Router			/api/v1/albums [get]
//	@Success		200	{object}	Response
func ListAlbums(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		user := entity.FindUser(username.(string))
		if user == nil {
			c.JSON(http.StatusInternalServerError, Error(400, "user not found"))
			return
		}

		albums, err := entity.ListAlbumsByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Error(400, "Albums get failed"))
			return
		}

		c.JSON(http.StatusOK, Success(albums))
	}

	router.GET("/albums", handler)
}

// UpdateAlbum 用于修改相册名称的处理函数
func UpdateAlbum(router *gin.RouterGroup, conf *config.Config) {
	handler := func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, Error(400, "User context not found"))
			return
		}

		user := entity.FindUser(username.(string))
		if user == nil {
			c.JSON(http.StatusInternalServerError, Error(400, "User not found"))
			return
		}

		// 获取相册ID
		albumIDStr := c.Param("id")
		albumID, err := strconv.ParseUint(albumIDStr, 10, 64) // 将字符串转换为 uint64
		if err != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid album ID"))
			return
		}

		var json struct {
			AlbumName string `json:"album_name"`
		}

		if c.BindJSON(&json) != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid request"))
			return
		}

		if entity.ExistsAlbum(json.AlbumName, user.Name) {
			c.JSON(http.StatusConflict, Error(400, "Album already exists"))
			return
		}

		// 检查相册是否存在
		album, err := entity.FindAlbumByID(uint(albumID))
		if err != nil {
			c.JSON(http.StatusBadRequest, Error(400, "Invalid request"))
			return
		}

		if album == nil {
			c.JSON(http.StatusNotFound, Error(404, "Album not found or does not belong to the user"))
			return
		}

		if err := album.Update(json.AlbumName); err != nil {
			c.JSON(http.StatusConflict, Error(400, "Album update failed"))
			return
		}

		c.JSON(http.StatusOK, Success(fmt.Sprintf("Album %s updated successfully", json.AlbumName)))
	}

	router.PUT("/albums/:id", handler)
}
