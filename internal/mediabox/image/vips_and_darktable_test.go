package image

import (
	"os"
	"testing"
)

func TestVips(t *testing.T) {
	source := "/Users/greene/Workspace/mediabox/mediabox-data/storage/originals/a63a6ca5-4ff3-44fc-9978-49df176cb7ad/a49c33ee-cd6b-449e-9acf-6539f7449f32/DSCF1336.JPG"
	target := "/Users/greene/Workspace/mediabox/DSCF1772.jpg"

	vips := NewVips()
	vips.Init()
	defer vips.Shutdown()

	err := vips.CreateThumbnail(source, target)
	t.Logf("%s", err)

	os.Remove(target)
}

func TestDarktable(t *testing.T) {
	source := "/Users/greene/Workspace/mediabox/mediabox-data/storage/originals/a63a6ca5-4ff3-44fc-9978-49df176cb7ad/a49c33ee-cd6b-449e-9acf-6539f7449f32/DSCF1772.RAF"
	target := "/Users/greene/Workspace/mediabox/DSCF1772.jpg"

	dt := NewDarktable()

	err := dt.CreateThumbnail(source, target)
	t.Logf("%s", err)

	os.Remove(target)
}
