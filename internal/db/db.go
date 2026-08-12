package db

import (
	"errors"
	"log"
	"sync"

	"github.com/numbereddev/zero-daemon/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	_db   *gorm.DB = nil
	_once sync.Once
)

var (
	ErrClientCreated    = errors.New("client already created")
	ErrClientNotCreated = errors.New("client not created")
)

func Init() error {
	ran := false
	_once.Do(func() {
		ran = true

		db, err := gorm.Open(sqlite.Open("zero.db"), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		_db = db
	})
	if ran {
		return ErrClientCreated
	}

	return nil
}

func Migrate() error {
	db, err := Client()
	if err != nil {
		return err
	}

	log.Print("Migrating database...")
	return db.AutoMigrate(&models.Service{})
}

func Client() (db *gorm.DB, err error) {
	if _db == nil {
		return nil, ErrClientNotCreated
	}

	return _db, nil
}
