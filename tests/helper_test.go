package tests

import (
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iWinston/qorm/tests/model"
)

type Config struct {
	Account   bool
	Pets      int
	Toys      int
	Company   bool
	Manager   bool
	Team      int
	Languages int
	Friends   int
	NamedPet  bool
}

func GetSimpleUser(name string, config Config) *model.SimpleUser {

	var (
		birthday = time.Now().Round(time.Second)
		user     = model.SimpleUser{
			Name:     name,
			Age:      18,
			Birthday: &birthday,
		}
	)
	if config.Account {
		user.Account = model.Account{Number: name + "_account"}
	}

	for i := 0; i < config.Pets; i++ {
		user.Pets = append(user.Pets, &model.Pet{Name: name + "_pet_" + strconv.Itoa(i+1)})
	}

	for i := 0; i < config.Toys; i++ {
		user.Toys = append(user.Toys, model.Toy{Name: name + "_toy_" + strconv.Itoa(i+1)})
	}

	if config.Company {
		user.Company = model.Company{Name: "company-" + name}
	}

	if config.Manager {
		user.Manager = GetUser(name+"_manager", Config{})
	}

	for i := 0; i < config.Team; i++ {
		user.Team = append(user.Team, *GetUser(name+"_team_"+strconv.Itoa(i+1), Config{}))
	}

	for i := 0; i < config.Languages; i++ {
		name := name + "_locale_" + strconv.Itoa(i+1)
		language := model.Language{Code: name, Name: name}
		user.Languages = append(user.Languages, language)
	}

	for i := 0; i < config.Friends; i++ {
		user.Friends = append(user.Friends, GetUser(name+"_friend_"+strconv.Itoa(i+1), Config{}))
	}

	if config.NamedPet {
		user.NamedPet = &model.Pet{Name: name + "_namepet"}
	}

	return &user
}

func GetUser(name string, config Config) *model.User {

	var (
		birthday = time.Now().Round(time.Second)
		user     = model.User{
			Name:     name,
			Age:      18,
			Birthday: &birthday,
		}
	)

	if config.Account {
		user.Account = model.Account{Number: name + "_account"}
	}

	for i := 0; i < config.Pets; i++ {
		user.Pets = append(user.Pets, &model.Pet{Name: name + "_pet_" + strconv.Itoa(i+1)})
	}

	for i := 0; i < config.Toys; i++ {
		user.Toys = append(user.Toys, model.Toy{Name: name + "_toy_" + strconv.Itoa(i+1)})
	}

	if config.Company {
		user.Company = model.Company{Name: "company-" + name}
	}

	if config.Manager {
		user.Manager = GetUser(name+"_manager", Config{})
	}

	for i := 0; i < config.Team; i++ {
		user.Team = append(user.Team, *GetUser(name+"_team_"+strconv.Itoa(i+1), Config{}))
	}

	for i := 0; i < config.Languages; i++ {
		name := name + "_locale_" + strconv.Itoa(i+1)
		language := model.Language{Code: name, Name: name}
		user.Languages = append(user.Languages, language)
	}

	for i := 0; i < config.Friends; i++ {
		user.Friends = append(user.Friends, GetUser(name+"_friend_"+strconv.Itoa(i+1), Config{}))
	}

	if config.NamedPet {
		user.NamedPet = &model.Pet{Name: name + "_namepet"}
	}

	return &user
}

func CheckPet(t *testing.T, pet model.Pet, expect model.Pet) {
	if pet.ID != 0 {
		var newPet model.Pet
		if err := DB.Where("id = ?", pet.ID).First(&newPet).Error; err != nil {
			t.Fatalf("errors happened when query: %v", err)
		} else {
			AssertObjEqual(t, newPet, pet, "ID", "UserID", "Name")
		}
	}

	AssertObjEqual(t, pet, expect, "ID", "UserID", "Name")

	AssertObjEqual(t, pet.Toy, expect.Toy, "ID", "Name", "OwnerID", "OwnerType")

	if expect.Toy.Name != "" && expect.Toy.OwnerType != "pets" {
		t.Errorf("toys's OwnerType, expect: %v, got %v", "pets", expect.Toy.OwnerType)
	}
}
func CheckSimpleUser(t *testing.T, user model.SimpleUser, expect model.SimpleUser) {
	if user.ID != 0 {
		var newUser model.SimpleUser
		if err := DB.Where("id = ?", user.ID).First(&newUser).Error; err != nil {
			t.Fatalf("errors happened when query: %v", err)
		} else {
			AssertObjEqual(t, newUser, user, "ID", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	}

	AssertObjEqual(t, user, expect, "ID", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")

	t.Run("Account", func(t *testing.T) {
		AssertObjEqual(t, user.Account, expect.Account, "ID", "UserID", "Number")

		if user.Account.Number != "" {
			if !user.Account.UserID.Valid {
				t.Errorf("Account's foreign key should be saved")
			} else {
				var account model.Account
				DB.First(&account, "user_id = ?", user.ID)
				AssertObjEqual(t, account, user.Account, "ID", "UserID", "Number")
			}
		}
	})

	t.Run("Pets", func(t *testing.T) {
		if len(user.Pets) != len(expect.Pets) {
			t.Fatalf("pets should equal, expect: %v, got %v", len(expect.Pets), len(user.Pets))
		}

		sort.Slice(user.Pets, func(i, j int) bool {
			return user.Pets[i].ID > user.Pets[j].ID
		})

		sort.Slice(expect.Pets, func(i, j int) bool {
			return expect.Pets[i].ID > expect.Pets[j].ID
		})

		for idx, pet := range user.Pets {
			if pet == nil || expect.Pets[idx] == nil {
				t.Errorf("pets#%v should equal, expect: %v, got %v", idx, expect.Pets[idx], pet)
			} else {
				CheckPet(t, *pet, *expect.Pets[idx])
			}
		}
	})

	t.Run("Toys", func(t *testing.T) {
		if len(user.Toys) != len(expect.Toys) {
			t.Fatalf("toys should equal, expect: %v, got %v", len(expect.Toys), len(user.Toys))
		}

		sort.Slice(user.Toys, func(i, j int) bool {
			return user.Toys[i].ID > user.Toys[j].ID
		})

		sort.Slice(expect.Toys, func(i, j int) bool {
			return expect.Toys[i].ID > expect.Toys[j].ID
		})

		for idx, toy := range user.Toys {
			if toy.OwnerType != "users" {
				t.Errorf("toys's OwnerType, expect: %v, got %v", "users", toy.OwnerType)
			}

			AssertObjEqual(t, toy, expect.Toys[idx], "ID", "Name", "OwnerID", "OwnerType")
		}
	})

	t.Run("Company", func(t *testing.T) {
		AssertObjEqual(t, user.Company, expect.Company, "ID", "Name")
	})

	t.Run("Manager", func(t *testing.T) {
		if user.Manager != nil {
			if user.ManagerID == nil {
				t.Errorf("Manager's foreign key should be saved")
			} else {
				var manager model.User
				DB.First(&manager, "id = ?", *user.ManagerID)
				AssertObjEqual(t, manager, user.Manager, "ID", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
			}
		} else if user.ManagerID != nil {
			t.Errorf("Manager should not be created for zero value, got: %+v", user.ManagerID)
		}
	})

	t.Run("Team", func(t *testing.T) {
		if len(user.Team) != len(expect.Team) {
			t.Fatalf("Team should equal, expect: %v, got %v", len(expect.Team), len(user.Team))
		}

		sort.Slice(user.Team, func(i, j int) bool {
			return user.Team[i].ID > user.Team[j].ID
		})

		sort.Slice(expect.Team, func(i, j int) bool {
			return expect.Team[i].ID > expect.Team[j].ID
		})

		for idx, team := range user.Team {
			AssertObjEqual(t, team, expect.Team[idx], "ID", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	})

	t.Run("Languages", func(t *testing.T) {
		if len(user.Languages) != len(expect.Languages) {
			t.Fatalf("Languages should equal, expect: %v, got %v", len(expect.Languages), len(user.Languages))
		}

		sort.Slice(user.Languages, func(i, j int) bool {
			return strings.Compare(user.Languages[i].Code, user.Languages[j].Code) > 0
		})

		sort.Slice(expect.Languages, func(i, j int) bool {
			return strings.Compare(expect.Languages[i].Code, expect.Languages[j].Code) > 0
		})
		for idx, language := range user.Languages {
			AssertObjEqual(t, language, expect.Languages[idx], "Code", "Name")
		}
	})

	t.Run("Friends", func(t *testing.T) {
		if len(user.Friends) != len(expect.Friends) {
			t.Fatalf("Friends should equal, expect: %v, got %v", len(expect.Friends), len(user.Friends))
		}

		sort.Slice(user.Friends, func(i, j int) bool {
			return user.Friends[i].ID > user.Friends[j].ID
		})

		sort.Slice(expect.Friends, func(i, j int) bool {
			return expect.Friends[i].ID > expect.Friends[j].ID
		})

		for idx, friend := range user.Friends {
			AssertObjEqual(t, friend, expect.Friends[idx], "ID", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	})
}

func CheckUser(t *testing.T, user model.User, expect model.User) {
	if user.ID != 0 {
		var newUser model.User
		if err := DB.Where("id = ?", user.ID).First(&newUser).Error; err != nil {
			t.Fatalf("errors happened when query: %v", err)
		} else {
			AssertObjEqual(t, newUser, user, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	}

	AssertObjEqual(t, user, expect, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")

	t.Run("Account", func(t *testing.T) {
		AssertObjEqual(t, user.Account, expect.Account, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "UserID", "Number")

		if user.Account.Number != "" {
			if !user.Account.UserID.Valid {
				t.Errorf("Account's foreign key should be saved")
			} else {
				var account model.Account
				DB.First(&account, "user_id = ?", user.ID)
				AssertObjEqual(t, account, user.Account, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "UserID", "Number")
			}
		}
	})

	t.Run("Pets", func(t *testing.T) {
		if len(user.Pets) != len(expect.Pets) {
			t.Fatalf("pets should equal, expect: %v, got %v", len(expect.Pets), len(user.Pets))
		}

		sort.Slice(user.Pets, func(i, j int) bool {
			return user.Pets[i].ID > user.Pets[j].ID
		})

		sort.Slice(expect.Pets, func(i, j int) bool {
			return expect.Pets[i].ID > expect.Pets[j].ID
		})

		for idx, pet := range user.Pets {
			if pet == nil || expect.Pets[idx] == nil {
				t.Errorf("pets#%v should equal, expect: %v, got %v", idx, expect.Pets[idx], pet)
			} else {
				CheckPet(t, *pet, *expect.Pets[idx])
			}
		}
	})

	t.Run("Toys", func(t *testing.T) {
		if len(user.Toys) != len(expect.Toys) {
			t.Fatalf("toys should equal, expect: %v, got %v", len(expect.Toys), len(user.Toys))
		}

		sort.Slice(user.Toys, func(i, j int) bool {
			return user.Toys[i].ID > user.Toys[j].ID
		})

		sort.Slice(expect.Toys, func(i, j int) bool {
			return expect.Toys[i].ID > expect.Toys[j].ID
		})

		for idx, toy := range user.Toys {
			if toy.OwnerType != "users" {
				t.Errorf("toys's OwnerType, expect: %v, got %v", "users", toy.OwnerType)
			}

			AssertObjEqual(t, toy, expect.Toys[idx], "ID", "CreatedAt", "UpdatedAt", "Name", "OwnerID", "OwnerType")
		}
	})

	t.Run("Company", func(t *testing.T) {
		AssertObjEqual(t, user.Company, expect.Company, "ID", "Name")
	})

	t.Run("Manager", func(t *testing.T) {
		if user.Manager != nil {
			if user.ManagerID == nil {
				t.Errorf("Manager's foreign key should be saved")
			} else {
				var manager model.User
				DB.First(&manager, "id = ?", *user.ManagerID)
				AssertObjEqual(t, manager, user.Manager, "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
			}
		} else if user.ManagerID != nil {
			t.Errorf("Manager should not be created for zero value, got: %+v", user.ManagerID)
		}
	})

	t.Run("Team", func(t *testing.T) {
		if len(user.Team) != len(expect.Team) {
			t.Fatalf("Team should equal, expect: %v, got %v", len(expect.Team), len(user.Team))
		}

		sort.Slice(user.Team, func(i, j int) bool {
			return user.Team[i].ID > user.Team[j].ID
		})

		sort.Slice(expect.Team, func(i, j int) bool {
			return expect.Team[i].ID > expect.Team[j].ID
		})

		for idx, team := range user.Team {
			AssertObjEqual(t, team, expect.Team[idx], "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	})

	t.Run("Languages", func(t *testing.T) {
		if len(user.Languages) != len(expect.Languages) {
			t.Fatalf("Languages should equal, expect: %v, got %v", len(expect.Languages), len(user.Languages))
		}

		sort.Slice(user.Languages, func(i, j int) bool {
			return strings.Compare(user.Languages[i].Code, user.Languages[j].Code) > 0
		})

		sort.Slice(expect.Languages, func(i, j int) bool {
			return strings.Compare(expect.Languages[i].Code, expect.Languages[j].Code) > 0
		})
		for idx, language := range user.Languages {
			AssertObjEqual(t, language, expect.Languages[idx], "Code", "Name")
		}
	})

	t.Run("Friends", func(t *testing.T) {
		if len(user.Friends) != len(expect.Friends) {
			t.Fatalf("Friends should equal, expect: %v, got %v", len(expect.Friends), len(user.Friends))
		}

		sort.Slice(user.Friends, func(i, j int) bool {
			return user.Friends[i].ID > user.Friends[j].ID
		})

		sort.Slice(expect.Friends, func(i, j int) bool {
			return expect.Friends[i].ID > expect.Friends[j].ID
		})

		for idx, friend := range user.Friends {
			AssertObjEqual(t, friend, expect.Friends[idx], "ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Name", "Age", "Birthday", "CompanyID", "ManagerID", "Active")
		}
	})
}
