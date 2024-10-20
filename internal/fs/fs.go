package fs

type FileType string

const (
	ImageARW = "arw"
	ImageRAF = "raf"
)

var RAWTypes = []FileType{
	ImageARW,
	ImageRAF,
}

func IsRAWData(ext FileType) bool {
	for _, v := range RAWTypes {
		if v == ext {
			return true
		}
	}
	return false
}
