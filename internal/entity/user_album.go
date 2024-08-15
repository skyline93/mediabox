package entity

type UserAlbum struct {
	UserID  uint
	AlbumID uint
}

func (UserAlbum) TableName() string {
	return "users_albums"
}
