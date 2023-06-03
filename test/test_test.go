package test

import (
	"github.com/iWinston/qorm"
	"github.com/iWinston/qorm/test/model"
	"log"
	"os"
)

var DB *qorm.DB

func init() {
	if db, err := OpenTestConnection(); err != nil {
		log.Printf("failed to connect database, got error %v", err)
		os.Exit(1)
	} else {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("failed to connect database, got error %v", err)
			os.Exit(1)
		}

		err = sqlDB.Ping()
		if err != nil {
			log.Printf("failed to ping sqlDB, got error %v", err)
			os.Exit(1)
		}

		allModels := []interface{}{&model.User{}, &model.Account{}, &model.Pet{}, &model.Company{}, &model.Toy{}, &model.Language{}, &model.Coupon{}, &model.CouponProduct{}, &model.Order{}, &model.Parent{}, &model.Child{}}
		DropTables(db, []interface{}{"user_friends", "user_speaks"})
		RunMigrations(db, allModels)

		if db.Dialector.Name() == "sqlite" {
			db.Exec("PRAGMA foreign_keys = ON")
		}
		DB = qorm.New(db)
	}
}
