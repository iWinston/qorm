package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type WhereParam struct {
	Name *string `where:"="`
}

type SimpleUser struct {
	gorm.Model
	Name      string
	Age       uint
	Birthday  *time.Time
	Account   SimpleAccount `gorm:"foreignKey:UserID"`
	Pets      []*SimplePet  `gorm:"foreignKey:UserID"`
	NamedPet  *SimplePet    `gorm:"foreignKey:UserID"`
	Toys      []SimpleToy   `gorm:"polymorphic:Owner"`
	CompanyID *int
	Company   SimpleCompany `gorm:"foreignKey:CompanyID"`
	ManagerID *uint
	Manager   *SimpleUser      `gorm:"foreignKey:ManagerID"`
	Team      []SimpleUser     `gorm:"foreignKey:ManagerID"`
	Languages []SimpleLanguage `gorm:"many2many:UserSpeak;foreignKey:ID;joinForeignKey:UserID;References:Code;JoinReferences:LanguageCode"`
	Friends   []*SimpleUser    `gorm:"many2many:user_friends;foreignKey:ID;joinForeignKey:UserID;References:ID;JoinReferences:FriendID"`
	Active    bool
}

func (i SimpleUser) TableName() string {
	return "users"
}

type SimpleAccount struct {
	gorm.Model
	UserID sql.NullInt64
	Number string
}

func (a *SimpleAccount) TableName() string {
	return "accounts"
}

type SimplePet struct {
	gorm.Model
	UserID *uint
	Name   string
	Toy    SimpleToy `gorm:"polymorphic:Owner;"`
}

func (s *SimplePet) TableName() string {
	return "pets"
}

type SimpleToy struct {
	gorm.Model
	Name      string
	OwnerID   string
	OwnerType string
}

func (s *SimpleToy) TableName() string {
	return "toys"
}

type SimpleCompany struct {
	ID   int
	Name string
}

func (c *SimpleCompany) TableName() string {
	return "companies"
}

type SimpleLanguage struct {
	Code string `gorm:"primarykey"`
	Name string
}

func (l *SimpleLanguage) TableName() string {
	return "languages"
}

type SimpleCoupon struct {
	ID                int                    `gorm:"primarykey; size:255"`
	AppliesToProducts []*SimpleCouponProduct `gorm:"foreignKey:CouponId;constraint:OnDelete:CASCADE"`
	AmountOff         uint32                 `gorm:"amount_off"`
	PercentOff        float32                `gorm:"percent_off"`
}

type SimpleCouponProduct struct {
	CouponId  int    `gorm:"primarykey;size:255"`
	ProductId string `gorm:"primarykey;size:255"`
	Desc      string
	CompanyID *int
	Company   SimpleCompany
}

type SimpleOrder struct {
	gorm.Model
	Num      string
	Coupon   *SimpleCoupon
	CouponID string
}

type SimpleParent struct {
	gorm.Model
	FavChildID uint
	FavChild   *SimpleChild
	Children   []*SimpleChild
}

type SimpleChild struct {
	gorm.Model
	Name     string
	ParentID *uint
	Parent   *SimpleParent
}
