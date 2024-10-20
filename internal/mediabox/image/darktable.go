package image

import (
	"fmt"
	"os/exec"
)

const (
	DefaultThumbnailWidth  = 0
	DefaultThumbnailHeight = 500
)

const DefaultDarktableBin = "darktable-cli"

type Darktable struct {
	Bin string
}

func NewDarktable() *Darktable {
	return &Darktable{Bin: DefaultDarktableBin}
}

func (d *Darktable) CreateThumbnail(source, target string) error {
	cmd := exec.Command(d.Bin, source, target, "--width", fmt.Sprintf("%d", DefaultThumbnailWidth), "--height", fmt.Sprintf("%d", DefaultThumbnailHeight))

	logger.Infof("run cmd: %s", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("err: %s, out: %s", err, out)
	}

	return nil
}
