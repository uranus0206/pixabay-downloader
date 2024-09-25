package dbmanager

import (
	// "gorm.io/driver/sqlite" // 基于 GGO 的 Sqlite 驱动

	"os"
	log "pixabay-downloader/pkg/qlogger"

	"github.com/glebarez/sqlite" // 纯 Go 实现的 SQLite 驱动, 详情参考： https://github.com/glebarez/sqlite
	"gorm.io/gorm"
)

type DbManager struct {
	gormDB *gorm.DB
}

var Dbm *DbManager

const (
	DbPixabayKey = "DbPixabayKey"
)

type DbPixabay struct {
	PRKey      string `gorm:"primaryKey"`
	CurrentTag string
	PageOffset uint64
}

func InitWithPath(path string) *DbManager {
	gormd, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	Dbm = &DbManager{
		gormDB: gormd,
	}

	return Dbm
}

func DeleteWithPath(path string) error {
	return os.Remove(path)
}

func (db *DbManager) CreateTable() {
	db.gormDB.AutoMigrate(&DbPixabay{})
}

func (db *DbManager) AddPixabayRecord(record *DbPixabay) {
	// Update if already exist in db, or create a new one
	// var found DbTopics
	var old DbPixabay
	res := db.gormDB.First(&old)
	log.Println("db search: ", res.Error)
	if res.Error == gorm.ErrRecordNotFound ||
		res.Error == gorm.ErrInvalidValue {
		res := db.gormDB.Create(record)
		log.Println("Add new: ", res.Error)
	} else if res.Error == nil {
		// Update
		res := db.gormDB.Updates(record)
		log.Println("Update exist: ", record, ", err: ", res.Error)
	} else {
		log.Errorln("DB search error: ", res.Error)
	}
}

func (db *DbManager) GetPixabayByKey() (DbPixabay, error) {
	var found DbPixabay

	res := db.gormDB.First(&found).Where("pr_key = ?", DbPixabayKey)

	log.Println("SearchByKey: ", res.Error, found)

	return found, res.Error
}

func (db *DbManager) DeletePixabay(record *DbPixabay) error {
	res := db.gormDB.Unscoped().Delete(&record)
	return res.Error
}
