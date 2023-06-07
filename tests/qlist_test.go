package tests

import (
	"testing"

	"github.com/iWinston/qorm/tests/model"
)

type Req struct {
	Name string `where:"=;Team.NamedPet.Name"`
}

func TestList(t *testing.T) {
	users := []model.SimpleUser{
		*GetUser("find", Config{}),
		*GetUser("find2", Config{}),
		*GetUser("find", Config{}),
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("List", func(t *testing.T) {
		req := &Req{
			"find",
		}
		res := []model.SimpleUser{}
		var total int64 = -1
		if err := DB.Model(&model.SimpleUser{}).Preload("Pets").QList(req, &res, &total).Error; err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			//CheckSimpleUser(t, res[0], users[0])
			//CheckSimpleUser(t, res[1], users[2])
		}
	})
}
