package tests

import (
	"testing"

	"github.com/iWinston/qorm/internal/tests/model"
)

type TestQParamReq struct {
	Name string `where:"=;Name"`
}

func TestQParam(t *testing.T) {
	users := []model.User{
		*GetUser("find", Config{}),
		*GetUser("find2", Config{}),
		*GetUser("find", Config{}),
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("QParam and QList at the same time", func(t *testing.T) {
		req := &TestQParamReq{
			"find",
		}
		res := []model.User{}
		var total int64 = -1
		if err := DB.Debug().Model(&model.User{}).QParam(req).QList(req, &res, &total).Err(); err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			CheckUser(t, res[0], users[0])
			CheckUser(t, res[1], users[2])
		}
	})
}
