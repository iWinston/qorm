package tests

import (
	"regexp"
	"testing"
	"time"

	"github.com/iWinston/qorm/tests/model"

	"github.com/jinzhu/now"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QCreateOne  is only support created by a *struct, this func   is essentially  copying the value of the param to model and execute create
func TestQCreateOne(t *testing.T) {
	m := GetSimpleUser("create", Config{})
	user := model.User{}
	// when we want to insert a record to users table
	// if we execute DB.Create(m),we will get a error that table 'qorm.simple_users' doesn't exist without specify that its table name is users.
	// if we execute DB.Model(&model.user{}).QCreateOne(m),we don't have to specify that its table name is users explicitly
	if results := DB.Model(&user).QCreateOne(m); results.Err() != nil {
		t.Fatalf(" happened when create: %v", results.Err())
	} else if results.RowsAffected != 1 {
		t.Fatalf("rows affected expects: %v, got %v", 1, results.RowsAffected)
	}

	if user.ID == 0 {
		t.Errorf("user's primary key should has value after create, got : %v", user.ID)
	}

	if user.CreatedAt.IsZero() {
		t.Errorf("user's created at should be not zero")
	}

	if user.UpdatedAt.IsZero() {
		t.Errorf("user's updated at should be not zero")
	}

	var newUser model.User
	if err := DB.Where("id = ?", user.ID).First(&newUser).Err(); err != nil {
		t.Fatalf("errors happened when query: %v", err)
	} else {
		CheckUser(t, user, newUser)
	}
}

func TestCreateWithAssociations(t *testing.T) {
	m := GetSimpleUser("create_with_associations", Config{
		Account:   true,
		Pets:      2,
		Toys:      3,
		Company:   true,
		Manager:   true,
		Team:      4,
		Languages: 3,
		Friends:   1,
	})
	user := model.User{}
	if err := DB.Model(&user).QCreateOne(m).Err(); err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}
	var user2 model.User
	if err := DB.Model(&model.User{}).Preload("Account").
		Preload("Pets").Preload("Toys").
		Preload("Company").Preload("Manager").
		Preload("Team").Preload("Languages").
		Preload("Friends").Find(&user2, "id = ?", user.ID).Err(); err != nil {
		t.Fatalf("errors happend where find: %v", err)
	}
	CheckUser(t, user, user2)
}

func TestPolymorphicHasOne(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		m := model.SimplePet{
			Name: "PolymorphicHasOne",
			Toy:  model.Toy{Name: "Toy-PolymorphicHasOne"},
		}
		pet := model.Pet{}
		if err := DB.Model(&pet).QCreateOne(&m).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}

		var pet2 model.Pet
		DB.Model(&model.Pet{}).Preload("Toy").Find(&pet2, "id = ?", pet.ID)
		CheckPet(t, pet, pet2)
	})
}

func TestCreateWithExistingTimestamp(t *testing.T) {
	m := model.SimpleUser{Name: "CreateUserExistingTimestamp"}
	curTime := now.MustParse("2016-01-01")
	m.CreatedAt = curTime
	m.UpdatedAt = curTime
	user := model.User{}
	if err := DB.Model(&user).QCreateOne(&m).Err(); err != nil {
		t.Fatalf("errors happened when qputOne err:%v", err)
	}

	AssertEqual(t, user.CreatedAt, curTime)
	AssertEqual(t, user.UpdatedAt, curTime)

	var newUser model.User
	DB.First(&newUser, user.ID)

	AssertEqual(t, newUser.CreatedAt, curTime)
	AssertEqual(t, newUser.UpdatedAt, curTime)
}

func TestCreateWithNowFuncOverride(t *testing.T) {
	m := model.SimpleUser{Name: "CreateUserTimestampOverride"}
	curTime := now.MustParse("2016-01-01")

	NEW := DB.Session(&gorm.Session{
		NowFunc: func() time.Time {
			return curTime
		},
	})
	user := model.User{}
	NEW.Model(&user).QCreateOne(&m)

	AssertEqual(t, user.CreatedAt, curTime)
	AssertEqual(t, user.UpdatedAt, curTime)

	var newUser model.User
	NEW.First(&newUser, user.ID)

	AssertEqual(t, newUser.CreatedAt, curTime)
	AssertEqual(t, newUser.UpdatedAt, curTime)
}

func TestCreateWithNoGORMPrimaryKey(t *testing.T) {
	type JoinTable struct {
		UserID   uint
		FriendID uint
	}

	if err := DB.Migrator().DropTable(&JoinTable{}); err != nil {
		t.Fatalf("errors happened when dropTable err:%v", err)
	}
	if err := DB.AutoMigrate(&JoinTable{}); err != nil {
		t.Errorf("no error should happen when auto migrate, but got %v", err)
	}

	jt := JoinTable{UserID: 1, FriendID: 2}
	if err := DB.Model(&JoinTable{}).QCreateOne(&jt).Err(); err != nil {
		t.Errorf("No error should happen when create a record without a GORM primary key. But in the database this primary key exists and is the union of 2 or more fields\n But got: %s", err)
	}
}

func TestSelectWithCreate(t *testing.T) {
	m := *GetSimpleUser("select_create", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	user := model.User{}
	if err := DB.Model(&user).Select("Account", "Toys", "Manager", "ManagerID", "Languages", "Name", "CreatedAt", "Age", "Active").QCreateOne(&m).Err(); err != nil {
		t.Fatalf("errors happened when qcreateOne err:%v", err)
	}

	var user2 model.User
	if err := DB.Preload("Account").Preload("Pets").Preload("Toys").
		Preload("Company").Preload("Manager").Preload("Team").
		Preload("Languages").Preload("Friends").First(&user2, user.ID).Err(); err != nil {
		t.Fatalf("errors happened when first err:%v", err)
	}

	user.Birthday = nil
	user.Pets = nil
	user.Company = model.Company{}
	user.Team = nil
	user.Friends = nil

	CheckUser(t, user, user2)
}

func TestOmitWithCreate(t *testing.T) {
	m := *GetSimpleUser("omit_create", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	user := model.User{}
	if err := DB.Model(&user).Omit("Account", "Toys", "Manager", "Birthday").QCreateOne(&m).Err(); err != nil {
		t.Fatalf("errors happened when qcreateOne err:%v", err)
	}

	var result model.User
	if err := DB.Preload("Account").
		Preload("Pets").Preload("Toys").
		Preload("Company").Preload("Manager").Preload("Team").
		Preload("Languages").Preload("Friends").First(&result, user.ID).Err(); err != nil {
		t.Fatalf("errors happened when first err:%v", err)
	}

	user.Birthday = nil
	user.Account = model.Account{}
	user.Toys = nil
	user.Manager = nil

	CheckUser(t, user, result)

	m2 := *GetSimpleUser("omit_create", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	user2 := model.User{}
	if err := DB.Model(&user2).Omit(clause.Associations).QCreateOne(&m2).Err(); err != nil {
		t.Fatalf("errors happened when qcreateOne err:%v", err)
	}

	var result2 model.User
	if err := DB.Model(&model.User{}).Preload(clause.Associations).First(&result2, user2.ID).Err(); err != nil {
		t.Fatalf("errors happened when first err:%v", err)
	}

	user2.Account = model.Account{}
	user2.Toys = nil
	user2.Manager = nil
	user2.Company = model.Company{}
	user2.Pets = nil
	user2.Team = nil
	user2.Languages = nil
	user2.Friends = nil

	CheckUser(t, user2, result2)
}

func TestCreateFromSubQuery(t *testing.T) {
	m := model.SimpleUser{Name: "jinzhu"}
	user := model.User{}
	DB.Model(&user).QCreateOne(&m)

	subQuery := DB.Table("users").Where("name=?", user.Name).Select("id").DB

	result := DB.Session(&gorm.Session{DryRun: true}).Model(&model.Pet{}).Create([]map[string]interface{}{
		{
			"name":    "cat",
			"user_id": gorm.Expr("(?)", DB.Table("(?) as tmp", subQuery).Select("@uid:=id").DB),
		},
		{
			"name":    "dog",
			"user_id": gorm.Expr("@uid"),
		},
	})
	println(result.Statement.SQL.String())
	if !regexp.MustCompile(`INSERT INTO .pets. \(.name.,.user_id.\) .*VALUES \(.+,\(SELECT @uid:=id FROM \(SELECT id FROM .users. WHERE name=.+\) as tmp\)\),\(.+,@uid\)`).MatchString(result.Statement.SQL.String()) {
		t.Errorf("invalid insert SQL, got %v", result.Statement.SQL.String())
	}
}

func TestCreateNilPointer(t *testing.T) {
	var user *model.SimpleUser

	err := DB.Model(&model.User{}).QCreateOne(user).Err()
	if err == nil || err != gorm.ErrInvalidValue {
		t.Fatalf("it is not ErrInvalidValue")
	}
}

func TestCreateWithAutoIncrementCompositeKey(t *testing.T) {
	type CompositeKeyProduct struct {
		ProductID    int `gorm:"primaryKey;autoIncrement:true;"` // primary key
		LanguageCode int `gorm:"primaryKey;"`                    // primary key
		Code         string
		Name         string
	}

	if err := DB.Migrator().DropTable(&CompositeKeyProduct{}); err != nil {
		t.Fatalf("failed to migrate, got error %v", err)
	}
	if err := DB.AutoMigrate(&CompositeKeyProduct{}); err != nil {
		t.Fatalf("failed to migrate, got error %v", err)
	}

	prod := &CompositeKeyProduct{
		LanguageCode: 56,
		Code:         "Code56",
		Name:         "ProductName56",
	}
	if err := DB.Model(prod).QCreateOne(&prod).Error; err != nil {
		t.Fatalf("failed to create, got error %v", err)
	}

	newProd := &CompositeKeyProduct{}
	if err := DB.First(&newProd).Error; err != nil {
		t.Fatalf("errors happened when query: %v", err)
	} else {
		AssertObjEqual(t, newProd, prod, "ProductID", "LanguageCode", "Code", "Name")
	}
}

func TestCreateOnConfilctWithDefalutNull(t *testing.T) {
	type OnConfilctUser struct {
		ID     string
		Name   string `gorm:"default:null"`
		Email  string
		Mobile string `gorm:"default:'133xxxx'"`
	}

	err := DB.Migrator().DropTable(&OnConfilctUser{})
	AssertEqual(t, err, nil)
	err = DB.AutoMigrate(&OnConfilctUser{})
	AssertEqual(t, err, nil)

	u := OnConfilctUser{
		ID:     "on-confilct-user-id",
		Name:   "on-confilct-user-name",
		Email:  "on-confilct-user-email",
		Mobile: "on-confilct-user-mobile",
	}
	err = DB.Model(&u).QCreateOne(&u).Error
	AssertEqual(t, err, nil)

	u.Name = "on-confilct-user-name-2"
	u.Email = "on-confilct-user-email-2"
	u.Mobile = ""
	if err = DB.Model(&u).Clauses(clause.OnConflict{UpdateAll: true}).QCreateOne(&u).Err(); err != nil {
		t.Fatalf("errors happened when qcreateOne err:%v", err)
	}
	AssertEqual(t, err, nil)

	var u2 OnConfilctUser
	if err = DB.Where("id = ?", u.ID).First(&u2).Err(); err != nil {
		t.Fatalf("errors happened when first err:%v", err)
	}
	AssertEqual(t, err, nil)
	AssertEqual(t, u2.Name, "on-confilct-user-name-2")
	AssertEqual(t, u2.Email, "on-confilct-user-email-2")
	AssertEqual(t, u2.Mobile, "133xxxx")
}
