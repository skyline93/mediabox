package entity

import (
	"sync"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/log"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	logger *logrus.Logger
	dbConn *DbConn
)

func init() {
	logger = log.NewLogger("db.log")
}

const (
	DriverSQLite3    = "sqlite3"
	DriverMySQL      = "mysql"
	DriverPostgreSQL = "postgresql"
)

type DbConn struct {
	Driver string
	Dsn    string
	db     *gorm.DB
	once   sync.Once
}

func (g *DbConn) Db() *gorm.DB {
	g.once.Do(g.Open)

	if g.db == nil {
		logger.Fatal("database not connected")
	}

	return g.db
}

func (g *DbConn) Open() {
	var dialector gorm.Dialector

	switch g.Driver {
	case DriverMySQL:
		dialector = mysql.Open(g.Dsn)
	case DriverPostgreSQL:
		dialector = postgres.Open(g.Dsn)
	default:
		if g.Dsn == "" {
			g.Dsn = "mediabox.db"
		}
		dialector = sqlite.Open(g.Dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logger.Fatal(err)
	}

	g.db = db
}

func InitDb(conf *config.Config) {
	dbConn = &DbConn{
		Driver: conf.DbDriver,
		Dsn:    conf.DbDsn,
	}

	logger.Infof("start auto migrate")
	err := dbConn.Db().AutoMigrate(
		&User{},
		&Album{},
		&Photo{},
	)
	if err != nil {
		logger.Fatalf("migrate db error, %s", err)
	}
}

func Db() *gorm.DB {
	if dbConn == nil {
		return nil
	}

	return dbConn.Db()
}

func UnscopedDb() *gorm.DB {
	return Db().Unscoped()
}
