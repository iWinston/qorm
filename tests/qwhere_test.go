package tests

import (
	"testing"

	"github.com/iWinston/qorm/tests/model"
)

func TestWhere(t *testing.T) {
	users := []model.SimpleUser{
		*GetUser("find", Config{}),
		*GetUser("find1", Config{}),
		*GetUser("find", Config{}),
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("First", func(t *testing.T) {
		var first model.SimpleUser
		name := "find1"
		if err := DB.Debug().QWhere(&model.WhereParam{Name: &name}).First(&first).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			CheckSimpleUser(t, first, users[1])
		}
	})
}
