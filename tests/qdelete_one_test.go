package tests

import (
	"errors"
	"testing"

	"github.com/iWinston/qorm/tests/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestDelete(t *testing.T) {
	users := []model.User{
		*GetUser("delete", Config{}),
		*GetUser("delete", Config{}),
		*GetUser("delete", Config{})}

	if err := DB.Create(&users).Err(); err != nil {
		t.Errorf("errors happened when create: %v", err)
	}

	for _, user := range users {
		if user.ID == 0 {
			t.Fatalf("user's primary key should has value after create, got : %v", user.ID)
		}
	}
	tmpUser := struct {
		ID uint `where:""`
	}{
		ID: users[1].ID,
	}
	if res := DB.Model(&model.User{}).QWhere(tmpUser).QDeleteOne(); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when delete: %v, affected: %v", res.Error, res.RowsAffected)
	}

	var result model.User
	if err := DB.Where("id = ?", users[1].ID).First(&result).Err(); err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", err)
	}

	for _, user := range []model.User{users[0], users[2]} {
		result = model.User{}
		if err := DB.Where("id = ?", user.ID).First(&result).Error; err != nil {
			t.Errorf("no error should returns when query %v, but got %v", user.ID, err)
		}
	}

	for _, user := range []model.User{users[0], users[2]} {
		result = model.User{}
		if err := DB.Where("id = ?", user.ID).First(&result).Error; err != nil {
			t.Errorf("no error should returns when query %v, but got %v", user.ID, err)
		}
	}

	if err := DB.Model(&users[0]).QDeleteOne().Err(); err != nil {
		t.Errorf("errors happened when delete: %v", err)
	}

	if err := DB.Model(&model.User{}).QDeleteOne().Err(); errors.Is(err, gorm.ErrMissingWhereClause) {
		t.Errorf("errors happened when delete: %v", err)
	}

	if err := DB.Where("id = ?", users[0].ID).First(&result).Err(); err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", err)
	}
}

func TestDeleteWithTable(t *testing.T) {
	type UserWithDelete struct {
		gorm.Model
		Name string
	}

	DB.Table("deleted_users").Migrator().DropTable(UserWithDelete{})
	DB.Table("deleted_users").AutoMigrate(UserWithDelete{})

	user := UserWithDelete{Name: "delete1"}
	DB.Table("deleted_users").Create(&user)

	var result UserWithDelete
	if err := DB.Table("deleted_users").First(&result).Error; err != nil {
		t.Errorf("failed to find deleted user, got error %v", err)
	}

	AssertEqual(t, result, user)

	if err := DB.Table("deleted_users").Model(&result).QDeleteOne().Error; err != nil {
		t.Errorf("failed to delete user, got error %v", err)
	}

	var result2 UserWithDelete
	if err := DB.Table("deleted_users").First(&result2, user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("should raise record not found error, but got error %v", err)
	}

	var result3 UserWithDelete
	if err := DB.Table("deleted_users").Unscoped().First(&result3, user.ID).Error; err != nil {
		t.Fatalf("failed to find record, got error %v", err)
	}

	if err := DB.Table("deleted_users").Unscoped().Model(&result).QDeleteOne().Error; err != nil {
		t.Errorf("failed to delete user with unscoped, got error %v", err)
	}

	var result4 UserWithDelete
	if err := DB.Table("deleted_users").Unscoped().First(&result4, user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("should raise record not found error, but got error %v", err)
	}
}

func TestDeleteWithAssociations(t *testing.T) {
	user := GetUser("delete_with_associations", Config{Account: true, Pets: 2, Toys: 4, Company: true, Manager: true, Team: 1, Languages: 1, Friends: 1})

	if err := DB.Create(user).Error; err != nil {
		t.Fatalf("failed to create user, got error %v", err)
	}

	//DB.QDeleteOne(m) will execute DB.Select(clause.Associations, "Pets.Toy").Take(m) before DB.Delete(m),so QDeleteOne not support by this wording which is DB.Select(clause.Associations,"Pets.Toy").QDeleteOne()
	if err := DB.Select(clause.Associations, "Pets.Toy").Delete(&user).Error; err != nil {
		t.Fatalf("failed to delete user, got error %v", err)
	}
	// company and  manager relationship is belongs to , so they will not be deleted
	for key, value := range map[string]int64{"Account": 1, "Pets": 2, "Toys": 4, "Company": 1, "Manager": 1, "Team": 1, "Languages": 0, "Friends": 0} {
		if count := DB.Unscoped().Model(&user).Association(key).Count(); count != value {
			t.Errorf("user's %v expects: %v, got %v", key, value, count)
		}
	}

	for key, value := range map[string]int64{"Account": 0, "Pets": 0, "Toys": 0, "Company": 1, "Manager": 1, "Team": 0, "Languages": 0, "Friends": 0} {
		if count := DB.Model(&user).Association(key).Count(); count != value {
			t.Errorf("user's %v expects: %v, got %v", key, value, count)
		}
	}
}

func TestDeleteAssociationsWithUnscoped(t *testing.T) {
	user := GetUser("unscoped_delete_with_associations", Config{Account: true, Pets: 2, Toys: 4, Company: true, Manager: true, Team: 1, Languages: 1, Friends: 1})

	if err := DB.Create(user).Error; err != nil {
		t.Fatalf("failed to create user, got error %v", err)
	}
	// QDeleteOne also not support this that reason has already been said in TestDeleteWithAssociations
	if err := DB.Unscoped().Select(clause.Associations, "Pets.Toy").Delete(&user).Error; err != nil {
		t.Fatalf("failed to delete user, got error %v", err)
	}

	for key, value := range map[string]int64{"Account": 0, "Pets": 0, "Toys": 0, "Company": 1, "Manager": 1, "Team": 0, "Languages": 0, "Friends": 0} {
		if count := DB.Unscoped().Model(&user).Association(key).Count(); count != value {
			t.Errorf("user's %v expects: %v, got %v", key, value, count)
		}
	}

	for key, value := range map[string]int64{"Account": 0, "Pets": 0, "Toys": 0, "Company": 1, "Manager": 1, "Team": 0, "Languages": 0, "Friends": 0} {
		if count := DB.Model(&user).Association(key).Count(); count != value {
			t.Errorf("user's %v expects: %v, got %v", key, value, count)
		}
	}
}
