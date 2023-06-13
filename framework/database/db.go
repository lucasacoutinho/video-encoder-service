package database

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
	"github.com/lucasacoutinho/video-encoder-service/domain"
)

type Database struct {
	DB            *gorm.DB
	DSN           string
	DBType        string
	Debug         bool
	AutoMigrateDB bool
	Env           string
}

func NewDB() *Database {
	return &Database{}
}

func NewDBTest() *gorm.DB {
	db := NewDB()
	db.Env = "test"
	db.DBType = "sqlite3"
	db.DSN = ":memory:"
	db.AutoMigrateDB = true
	db.Debug = true

	conn, err := db.Connect()

	if err != nil {
		log.Fatalf("Test db error %v", err)
	}

	return conn
}

func (d *Database) Connect() (*gorm.DB, error) {
	var err error

	d.DB, err = gorm.Open(d.DBType, d.DSN)

	if err != nil {
		return nil, err
	}

	if d.Debug {
		d.DB.LogMode(true)
	}

	if d.AutoMigrateDB {
		d.DB.AutoMigrate(&domain.Video{}, &domain.Job{})
		d.DB.Model(domain.Job{}).AddForeignKey("video_id", "videos (id)", "CASCADE", "CASCADE")
	}

	return d.DB, nil
}
