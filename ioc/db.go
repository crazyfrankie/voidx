package ioc

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(conf.GetConf().Postgre.DSN,
		os.Getenv("PG_HOST"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DB"),
		os.Getenv("PG_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(conf.GetConf().Postgre.PoolSize)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.GetConf().Postgre.PoolMaxTime))

	if err := entity.AutoMigrate(db); err != nil {
		panic(err)
	}

	return db
}
