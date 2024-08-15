package mediabox

import (
	"github.com/skyline93/mediabox/internal/log"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

func init() {
	logger = log.NewLogger("server.log")
}
