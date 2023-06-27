package tests

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/iWinston/qorm/tests/model"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

// this func   is essentially  copying the value of the param to model and execute update
func TestUpdate(t *testing.T) {
	var (
		users = []*model.User{
			GetUser("update-1", Config{}),
			GetUser("update-2", Config{}),
			GetUser("update-3", Config{}),
		}
		user          = users[1]
		lastUpdatedAt time.Time
	)

	checkUpdatedAtChanged := func(name string, n time.Time) {
		if n.UnixNano() == lastUpdatedAt.UnixNano() {
			t.Errorf("%v: user's updated at should be changed, but got %v, was %v", name, n, lastUpdatedAt)
		}
		lastUpdatedAt = n
	}

	checkOtherData := func(name string) {
		var first, last model.User
		if err := DB.Where("id = ?", users[0].ID).First(&first).Err(); err != nil {
			t.Errorf("errors happened when query before user: %v", err)
		}
		CheckUser(t, first, *users[0])

		if err := DB.Where("id = ?", users[2].ID).First(&last).Err(); err != nil {
			t.Errorf("errors happened when query after user: %v", err)
		}
		CheckUser(t, last, *users[2])
	}

	if err := DB.Create(&users).Err(); err != nil {
		t.Fatalf("errors happened when create: %v", err)
	} else if user.ID == 0 {
		t.Fatalf("user's primary value should not zero, %v", user.ID)
	} else if user.UpdatedAt.IsZero() {
		t.Fatalf("user's updated at should not zero, %v", user.UpdatedAt)
	}
	lastUpdatedAt = user.UpdatedAt

	if err := DB.Model(user).QPatchOne(&struct {
		Age int
	}{Age: 10}).Err(); err != nil {
		t.Errorf("errors happened when update: %v", err)
	} else if user.Age != 10 {
		t.Errorf("Age should equals to 10, but got %v", user.Age)
	}
	checkUpdatedAtChanged("Update", user.UpdatedAt)
	checkOtherData("Update")

	var result model.User
	if err := DB.Where("id = ?", user.ID).First(&result).Err(); err != nil {
		t.Errorf("errors happened when query: %v", err)
	} else {
		CheckUser(t, result, *user)
	}

	values := map[string]interface{}{"Active": true, "age": 5}
	if res := DB.Model(user).QPatchOne(values); res.Err() != nil {
		t.Errorf("errors happened when update: %v", res.Err())
	} else if res.RowsAffected != 1 {
		t.Errorf("rows affected should be 1, but got : %v", res.RowsAffected)
	} else if user.Age != 5 {
		t.Errorf("Age should equals to 5, but got %v", user.Age)
	} else if user.Active != true {
		t.Errorf("Active should be true, but got %v", user.Active)
	}
	checkUpdatedAtChanged("Updates with map", user.UpdatedAt)
	checkOtherData("Updates with map")

	var result2 model.User
	if err := DB.Where("id = ?", user.ID).First(&result2).Err(); err != nil {
		t.Errorf("errors happened when query: %v", err)
	} else {
		CheckUser(t, result2, *user)
	}

	user.Active = false
	user.Age = 1
	if err := DB.Save(user).Err(); err != nil {
		t.Errorf("errors happened when update: %v", err)
	} else if user.Age != 1 {
		t.Errorf("Age should equals to 1, but got %v", user.Age)
	} else if user.Active != false {
		t.Errorf("Active should equals to false, but got %v", user.Active)
	}
	checkUpdatedAtChanged("Save", user.UpdatedAt)
	checkOtherData("Save")

	var result4 model.User
	if err := DB.Where("id = ?", user.ID).First(&result4).Err(); err != nil {
		t.Errorf("errors happened when query: %v", err)
	} else {
		CheckUser(t, result4, *user)
	}

	if rowsAffected := DB.Model(&result4).Where("age > 0").QPatchOne(&struct {
		Name string
	}{Name: "zhangsan"}).RowsAffected; rowsAffected != 1 {
		t.Errorf("should only update one record, but got %v", rowsAffected)
	}
}

func TestSelectWithUpdate(t *testing.T) {
	user := *GetUser("select_update", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	DB.Create(&user)

	var result model.User
	DB.First(&result, user.ID)

	user2 := *GetUser("select_update_new", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	result.Name = user2.Name
	result.Age = 50
	result.Account = user2.Account
	result.Pets = user2.Pets
	result.Toys = user2.Toys
	result.Company = user2.Company
	result.Manager = user2.Manager
	result.Team = user2.Team
	result.Languages = user2.Languages
	result.Friends = user2.Friends

	DB.Select("Name", "Account", "Toys", "Manager", "ManagerID", "Languages").Save(&result)

	var result2 model.User
	DB.Preload("Account").Preload("Pets").Preload("Toys").Preload("Company").Preload("Manager").Preload("Team").Preload("Languages").Preload("Friends").First(&result2, user.ID)

	result.Languages = append(user.Languages, result.Languages...)
	result.Toys = append(user.Toys, result.Toys...)

	sort.Slice(result.Languages, func(i, j int) bool {
		return strings.Compare(result.Languages[i].Code, result.Languages[j].Code) > 0
	})

	sort.Slice(result.Toys, func(i, j int) bool {
		return result.Toys[i].ID < result.Toys[j].ID
	})

	sort.Slice(result2.Languages, func(i, j int) bool {
		return strings.Compare(result2.Languages[i].Code, result2.Languages[j].Code) > 0
	})

	sort.Slice(result2.Toys, func(i, j int) bool {
		return result2.Toys[i].ID < result2.Toys[j].ID
	})
	resultId := result.ID
	AssertObjEqual(t, result2, result, "Name", "Account", "Toys", "Manager", "ManagerID", "Languages")
	// if  lack of the condition which is id =1 , where use updates it will lose the condition id = 1 , because of using select
	//DB.Model(&result).Select("Name", "Age").Where("id = ?", result.ID).QPatchOne(&model.User{Name: "update_with_select"})
	DB.Model(&result).Where("id = ?", result.ID).Select("Name", "Age").QPatchOne(&model.User{Name: "update_with_select"})
	if result.Age != 0 || result.Name != "update_with_select" {
		t.Fatalf("Failed to update struct with select, got %+v", result)
	}
	AssertObjEqual(t, result, user, "UpdatedAt")

	var result3 model.User
	DB.First(&result3, resultId)
	AssertObjEqual(t, result, result3, "Name", "Age", "UpdatedAt")

	DB.Model(&result).Where("id = ?", resultId).Select("Name", "Age", "UpdatedAt").QPatchOne(&model.User{Name: "update_with_select"})

	if utils.AssertEqual(result.UpdatedAt, user.UpdatedAt) {
		t.Fatalf("Update struct should update UpdatedAt, was %+v, got %+v", result.UpdatedAt, user.UpdatedAt)
	}

	AssertObjEqual(t, result, model.User{Name: "update_with_select"}, "Name", "Age")
}

func TestSelectWithUpdateWithMap(t *testing.T) {
	user := *GetUser("select_update_map", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	DB.Create(&user)

	var result model.User
	DB.First(&result, user.ID)

	user2 := *GetUser("select_update_map_new", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	updateValues := map[string]interface{}{
		"Name":      user2.Name,
		"Age":       50,
		"Account":   user2.Account,
		"Pets":      user2.Pets,
		"Toys":      user2.Toys,
		"Company":   user2.Company,
		"Manager":   user2.Manager,
		"Team":      user2.Team,
		"Languages": user2.Languages,
		"Friends":   user2.Friends,
	}
	name := result.Name
	// QPatchOne 会将param 复制到model,然后update，因此result的name会发生变化
	if err := DB.Model(&result).Omit("name", "updated_at").QPatchOne(&updateValues).Err(); err != nil {
		t.Fatalf("errors happened when qpatchone err:%v", err)
	}

	var result2 model.User
	DB.Preload("Account").Preload("Pets").Preload("Toys").Preload("Company").Preload("Manager").Preload("Team").Preload("Languages").Preload("Friends").First(&result2, user.ID)

	result.Languages = append(user.Languages, result.Languages...)
	result.Toys = append(user.Toys, result.Toys...)

	sort.Slice(result.Languages, func(i, j int) bool {
		return strings.Compare(result.Languages[i].Code, result.Languages[j].Code) > 0
	})

	sort.Slice(result.Toys, func(i, j int) bool {
		return result.Toys[i].ID < result.Toys[j].ID
	})

	sort.Slice(result2.Languages, func(i, j int) bool {
		return strings.Compare(result2.Languages[i].Code, result2.Languages[j].Code) > 0
	})

	sort.Slice(result2.Toys, func(i, j int) bool {
		return result2.Toys[i].ID < result2.Toys[j].ID
	})

	AssertObjEqual(t, result2, result, "Account", "Toys", "Manager", "ManagerID", "Languages")
	if !(result.Name != name && result2.Name == name) {
		t.Fatalf("name is not equal")
	}
}

func TestWithUpdateWithInvalidMap(t *testing.T) {
	user := *GetUser("update_with_invalid_map", Config{})
	DB.Create(&user)

	if err := DB.Model(&user).QPatchOne(map[string]string{"name": "jinzhu"}).Err(); err != nil {
		fmt.Println(err)
		t.Errorf("should returns error for unsupported updating data")
	}
}

func TestOmitWithUpdate(t *testing.T) {
	user := *GetUser("omit_update", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	DB.Create(&user)

	var result model.User
	DB.First(&result, user.ID)

	user2 := *GetSimpleUser("omit_update_new", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	result.Name = user2.Name
	result.Age = 50
	result.Account = user2.Account
	result.Pets = user2.Pets
	result.Toys = user2.Toys
	result.Company = user2.Company
	result.Manager = user2.Manager
	result.Team = user2.Team
	result.Languages = user2.Languages
	result.Friends = user2.Friends

	updateResult := model.User{}
	if err := DB.Model(&updateResult).Omit("Name", "Account", "Toys", "Manager", "ManagerID", "Languages").QPatchOne(&result).Err(); err != nil {
		t.Fatalf("errors happened when qpatchone err:%v", err)
	}
	var result2 model.User
	DB.Preload("Account").Preload("Pets").Preload("Toys").Preload("Company").Preload("Manager").Preload("Team").Preload("Languages").Preload("Friends").First(&result2, user.ID)

	updateResult.Pets = append(user.Pets, updateResult.Pets...)
	updateResult.Team = append(user.Team, updateResult.Team...)
	updateResult.Friends = append(user.Friends, updateResult.Friends...)

	sort.Slice(updateResult.Pets, func(i, j int) bool {
		return updateResult.Pets[i].ID < updateResult.Pets[j].ID
	})
	sort.Slice(updateResult.Team, func(i, j int) bool {
		return updateResult.Team[i].ID < updateResult.Team[j].ID
	})
	sort.Slice(updateResult.Friends, func(i, j int) bool {
		return updateResult.Friends[i].ID < updateResult.Friends[j].ID
	})
	sort.Slice(result2.Pets, func(i, j int) bool {
		return result2.Pets[i].ID < result2.Pets[j].ID
	})
	sort.Slice(result2.Team, func(i, j int) bool {
		return result2.Team[i].ID < result2.Team[j].ID
	})
	sort.Slice(result2.Friends, func(i, j int) bool {
		return result2.Friends[i].ID < result2.Friends[j].ID
	})

	AssertObjEqual(t, result2, updateResult, "Age", "Pets", "Company", "CompanyID", "Team", "Friends")
}

func TestOmitWithUpdateWithMap(t *testing.T) {
	user := *GetUser("omit_update_map", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	DB.Create(&user)

	var result model.User
	DB.First(&result, user.ID)

	user2 := *GetUser("omit_update_map_new", Config{Account: true, Pets: 3, Toys: 3, Company: true, Manager: true, Team: 3, Languages: 3, Friends: 4})
	updateValues := map[string]interface{}{
		"Name":      user2.Name,
		"Age":       50,
		"Account":   user2.Account,
		"Pets":      user2.Pets,
		"Toys":      user2.Toys,
		"Company":   user2.Company,
		"Manager":   user2.Manager,
		"Team":      user2.Team,
		"Languages": user2.Languages,
		"Friends":   user2.Friends,
	}

	if err := DB.Model(&result).Omit("Name", "Account", "Toys", "Manager", "ManagerID", "Languages").QPatchOne(updateValues).Err(); err != nil {
		t.Fatalf("errors happened when qpatchone err:%v", err)
	}

	var result2 model.User
	DB.Preload("Account").Preload("Pets").Preload("Toys").Preload("Company").Preload("Manager").Preload("Team").Preload("Languages").Preload("Friends").First(&result2, user.ID)

	result.Pets = append(user.Pets, result.Pets...)
	result.Team = append(user.Team, result.Team...)
	result.Friends = append(user.Friends, result.Friends...)

	sort.Slice(result.Pets, func(i, j int) bool {
		return result.Pets[i].ID < result.Pets[j].ID
	})
	sort.Slice(result.Team, func(i, j int) bool {
		return result.Team[i].ID < result.Team[j].ID
	})
	sort.Slice(result.Friends, func(i, j int) bool {
		return result.Friends[i].ID < result.Friends[j].ID
	})
	sort.Slice(result2.Pets, func(i, j int) bool {
		return result2.Pets[i].ID < result2.Pets[j].ID
	})
	sort.Slice(result2.Team, func(i, j int) bool {
		return result2.Team[i].ID < result2.Team[j].ID
	})
	sort.Slice(result2.Friends, func(i, j int) bool {
		return result2.Friends[i].ID < result2.Friends[j].ID
	})

	AssertObjEqual(t, result2, result, "Age", "Pets", "Company", "CompanyID", "Team", "Friends")
}

func TestUpdatesWithBlankValues(t *testing.T) {
	user := *GetUser("updates_with_blank_value", Config{})
	DB.Save(&user)

	var user2 model.User
	DB.Model(&user2).Where("id = ?", user.ID).QPatchOne(&model.User{Age: 100})

	var result model.User
	DB.First(&result, user.ID)

	if result.Name != user.Name || result.Age != 100 {
		t.Errorf("user's name should not be updated")
	}
}

func TestUpdatesTableWithIgnoredValues(t *testing.T) {
	type ElementWithIgnoredField struct {
		Id           int64
		Value        string
		IgnoredField int64 `gorm:"-"`
	}
	DB.Migrator().DropTable(&ElementWithIgnoredField{})
	DB.AutoMigrate(&ElementWithIgnoredField{})

	elem := ElementWithIgnoredField{Value: "foo", IgnoredField: 10}
	DB.Save(&elem)

	DB.Model(&ElementWithIgnoredField{}).
		Where("id = ?", elem.Id).
		QPatchOne(&ElementWithIgnoredField{Value: "bar", IgnoredField: 100})

	var result ElementWithIgnoredField
	if err := DB.First(&result, elem.Id).Err(); err != nil {
		t.Errorf("error getting an element from database: %s", err.Error())
	}

	if result.IgnoredField != 0 {
		t.Errorf("element's ignored field should not be updated")
	}
}

func TestUpdateFromSubQuery(t *testing.T) {
	user := *GetUser("update_from_sub_query", Config{Company: true})
	if err := DB.Create(&user).Err(); err != nil {
		t.Errorf("failed to create user, got error: %v", err)
	}

	if err := DB.Model(&user).Update("name", DB.Model(&model.Company{}).Select("name").Where("companies.id = users.company_id").DB).Err(); err != nil {
		t.Errorf("failed to update with sub query, got error %v", err)
	}

	var result model.User
	DB.First(&result, user.ID)

	if result.Name != user.Company.Name {
		t.Errorf("name should be %v, but got %v", user.Company.Name, result.Name)
	}

	DB.Model(&user.Company).Update("Name", "new company name")
	if err := DB.Table("users").Where("1 = 1").Update("name", DB.Table("companies").Select("name").Where("companies.id = users.company_id").DB).Err(); err != nil {
		t.Errorf("failed to update with sub query, got error %v", err)
	}

	DB.First(&result, user.ID)
	if result.Name != "new company name" {
		t.Errorf("name should be %v, but got %v", user.Company.Name, result.Name)
	}
}

func TestIdempotentSave(t *testing.T) {
	create := model.Company{
		Name: "company_idempotent",
	}
	DB.Create(&create)

	var company model.Company
	if err := DB.Find(&company, "id = ?", create.ID).Err(); err != nil {
		t.Fatalf("failed to find created company, got err: %v", err)
	}

	if err := DB.Model(&model.Company{}).QPutOne(&company).Err(); err != nil || company.ID != create.ID {
		t.Errorf("failed to save company, got err: %v", err)
	}
	if err := DB.Model(&model.Company{}).QPutOne(&company).Err(); err != nil || company.ID != create.ID {
		t.Errorf("failed to save company, got err: %v", err)
	}
}

func TestSave(t *testing.T) {
	user := *GetUser("save", Config{})
	DB.Create(&user)

	if err := DB.First(&model.User{}, "name = ?", "save").Err(); err != nil {
		t.Fatalf("failed to find created user")
	}

	user.Name = "save2"
	DB.Model(&model.User{}).QPutOne(&user)

	var result model.User
	if err := DB.First(&result, "name = ?", "save2").Err(); err != nil || result.ID != user.ID {
		t.Fatalf("failed to find updated user")
	}

	user2 := *GetUser("save2", Config{})
	DB.Create(&user2)

	time.Sleep(time.Second)
	user1UpdatedAt := result.UpdatedAt
	user2UpdatedAt := user2.UpdatedAt
	users := []*model.User{&result, &user2}
	DB.Model(&model.User{}).Save(&users)

	if user1UpdatedAt.Format(time.RFC1123Z) == result.UpdatedAt.Format(time.RFC1123Z) {
		t.Fatalf("user's updated at should be changed, expects: %+v, got: %+v", user1UpdatedAt, result.UpdatedAt)
	}

	if user2UpdatedAt.Format(time.RFC1123Z) == user2.UpdatedAt.Format(time.RFC1123Z) {
		t.Fatalf("user's updated at should be changed, expects: %+v, got: %+v", user2UpdatedAt, user2.UpdatedAt)
	}

	DB.First(&result)
	if user1UpdatedAt.Format(time.RFC1123Z) == result.UpdatedAt.Format(time.RFC1123Z) {
		t.Fatalf("user's updated at should be changed after reload, expects: %+v, got: %+v", user1UpdatedAt, result.UpdatedAt)
	}

	DB.First(&user2)
	if user2UpdatedAt.Format(time.RFC1123Z) == user2.UpdatedAt.Format(time.RFC1123Z) {
		t.Fatalf("user2's updated at should be changed after reload, expects: %+v, got: %+v", user2UpdatedAt, user2.UpdatedAt)
	}

	dryDB := DB.Session(&gorm.Session{DryRun: true})
	stmt := dryDB.Model(&model.User{}).QPutOne(&user).Statement
	if !regexp.MustCompile(`.users.\..deleted_at. IS NULL`).MatchString(stmt.SQL.String()) {
		t.Fatalf("invalid updating SQL, got %v", stmt.SQL.String())
	}

	dryDB = DB.Session(&gorm.Session{DryRun: true})
	stmt = dryDB.Unscoped().Model(&model.User{}).QPutOne(&user).Statement
	if !regexp.MustCompile(`WHERE .id. = [^ ]+$`).MatchString(stmt.SQL.String()) {
		t.Fatalf("invalid updating SQL, got %v", stmt.SQL.String())
	}

	user3 := *GetUser("save3", Config{})
	DB.Create(&user3)

	if err := DB.First(&model.User{}, "name = ?", "save3").Err(); err != nil {
		t.Fatalf("failed to find created user")
	}

	user3.Name = "save3_"
	if err := DB.Model(&model.User{Model: user3.Model}).QPutOne(&user3).Err(); err != nil {
		t.Fatalf("failed to save user, got %v", err)
	}

	var result2 model.User
	if err := DB.First(&result2, "name = ?", "save3_").Err(); err != nil || result2.ID != user3.ID {
		t.Fatalf("failed to find updated user, got %v", err)
	}

	if err := DB.Model(model.User{Model: user3.Model}).Save(&struct {
		gorm.Model
		Placeholder string
		Name        string
	}{
		Model:       user3.Model,
		Placeholder: "placeholder",
		Name:        "save3__",
	}).Err(); err != nil {
		t.Fatalf("failed to update user, got %v", err)
	}

	var result3 model.User
	if err := DB.First(&result3, "name = ?", "save3__").Err(); err != nil || result3.ID != user3.ID {
		t.Fatalf("failed to find updated user")
	}
}

func TestSaveWithPrimaryValue(t *testing.T) {
	lang := model.Language{Code: "save", Name: "save"}
	if result := DB.Save(&lang); result.RowsAffected != 1 {
		t.Errorf("should create language, rows affected: %v", result.RowsAffected)
	}

	var result model.Language
	DB.First(&result, "code = ?", "save")
	AssertEqual(t, result, lang)

	lang.Name = "save name2"
	if result := DB.Model(&model.Language{}).QPutOne(&lang); result.RowsAffected != 1 {
		t.Errorf("should update language")
	}

	var result2 model.Language
	DB.First(&result2, "code = ?", "save")
	AssertEqual(t, result2, lang)

	DB.Table("langs").Migrator().DropTable(&model.Language{})
	DB.Table("langs").AutoMigrate(&model.Language{})

	if err := DB.Table("langs").Save(&lang).Err(); err != nil {
		t.Errorf("no error should happen when creating data, but got %v", err)
	}

	var result3 model.Language
	if err := DB.Table("langs").First(&result3, "code = ?", lang.Code).Err(); err != nil || result3.Name != lang.Name {
		t.Errorf("failed to find created record, got error: %v, result: %+v", err, result3)
	}

	lang.Name += "name2"
	if err := DB.Table("langs").Save(&lang).Err(); err != nil {
		t.Errorf("no error should happen when creating data, but got %v", err)
	}

	var result4 model.Language
	if err := DB.Table("langs").First(&result4, "code = ?", lang.Code).Err(); err != nil || result4.Name != lang.Name {
		t.Errorf("failed to find created record, got error: %v, result: %+v", err, result4)
	}
}

func TestUpdateWithDiffSchema(t *testing.T) {
	user := GetUser("update-diff-schema-1", Config{})
	DB.Create(&user)

	type UserTemp struct {
		Name string
	}

	err := DB.Model(&user).Where("id = ?", user.ID).QPatchOne(&UserTemp{Name: "update-diff-schema-2"}).Err()
	AssertEqual(t, err, nil)
	AssertEqual(t, "update-diff-schema-2", user.Name)
}
