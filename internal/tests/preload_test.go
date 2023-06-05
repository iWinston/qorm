package tests

import (
	"fmt"
	"testing"

	"github.com/iWinston/qorm/internal/tests/model"

	"gorm.io/gorm"
)

func TestPreload(t *testing.T) {
	cfg := Config{Account: true,
		Pets:      2,
		Toys:      3,
		Company:   true,
		Manager:   true,
		Team:      2,
		Languages: 3,
		Friends:   3,
		NamedPet:  true}

	names := []string{"find1", "find2", "find3"}
	users := make([]model.User, 0, len(names))
	for _, name := range names {
		users = append(users, *GetUser(name, cfg))
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("JoinQuery", func(t *testing.T) {
		var userMsg []UserCMsg
		if err := DB.Debug().Model(&model.User{}).Where("`users`.`name` in (?)", names).
			Find(&userMsg).Err(); err != nil {
			panic(err)
		}
		for _, msg := range userMsg {
			fmt.Printf("%+v\n", msg)
		}
	})
	t.Run("PreloadQuery", func(t *testing.T) {
		var userMsg []UserMsg
		if err := DB.Debug().Model(&model.User{}).Where("`users`.`name` in (?)", names).
			Preload("Pets", func(db *gorm.DB) *gorm.DB { return db.Model(&model.Pet{}) }).
			Preload("Manager", func(db *gorm.DB) *gorm.DB { return db.Model(&model.User{}) }).
			Find(&userMsg).Err(); err != nil {
			panic(err)
		}
		for _, msg := range userMsg {
			fmt.Printf("%+v\n", msg)
		}
	})
}

type UserCMsg struct {
	Id          uint
	Name        string
	Age         uint
	CompanyName string `select:"Company.Name"`
}

type UserMsg struct {
	Id          uint
	Name        string
	Age         uint
	CompanyName string `select:"Company.Name"` // 放开注释会查询报空指针异常,join插件调用了两次，第二次报空指针
	ManagerID   uint
	Manager     SimpleManager `gorm:"foreignKey:ManagerID"`
	Pets        []SimplePet   `gorm:"foreignKey:UserID"`
}

type SimplePet struct {
	Id     uint
	UserID uint
	Name   string
}

type SimpleManager struct {
	Id   uint
	Name string
}
