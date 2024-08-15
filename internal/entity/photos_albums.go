package entity

type PhotoAlbum struct {
	PhotoID uint
	AlbumID uint
}

func (PhotoAlbum) TableName() string {
	return "photos_albums"
}
