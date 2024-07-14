package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/guobinqiu/appdeployer/api/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AppDeployerDB struct {
	db     *gorm.DB
	mu     *sync.Mutex
	dbPath string
}

func NewAppDeployerDB(dbPath string) *AppDeployerDB {
	return &AppDeployerDB{
		dbPath: dbPath,
		mu:     &sync.Mutex{},
	}
}

func (s *AppDeployerDB) GetDB() *gorm.DB {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Dir(s.dbPath)
	os.MkdirAll(dir, os.ModePerm)
	if s.db == nil {
		db, err := gorm.Open(sqlite.Open(s.dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(fmt.Sprintf("Init database connection error: %s", err.Error()))
		}
		s.db = db
	}
	return s.db
}

func (s *AppDeployerDB) Close() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		sqlDB.Close()
		s.db = nil
	}
}

func (s *AppDeployerDB) Migrate() {
	m := gormigrate.New(s.GetDB(), gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202204181040",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&model.User{},
					&model.Client{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&model.User{},
					&model.Client{},
				)
			},
		},
		{
			ID: "202204181041",
			Migrate: func(tx *gorm.DB) error {
				client := model.Client{
					ClientKey:    "10412844c46211ec83d58c8590a958ee",
					ClientSecret: HashPassword("222222222222222Aa*"),
				}
				tx.Create(&client)
				tx.Create(&model.User{
					Username: "admin",
					Password: HashPassword("111111111111Aa*"),
					ClientID: client.ID,
				})
				return nil
			},
		},
	})
	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Println("Migration did run successfully")
}
