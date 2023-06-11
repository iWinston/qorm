package tests

import (
	"testing"

	"github.com/iWinston/qorm/tests/model"
)

type TestQParamReq struct {
	Name string `where:"=;Name"`
}

func TestQParam(t *testing.T) {
	users := []model.SimpleUser{
		*GetSimpleUser("find", Config{}),
		*GetSimpleUser("find2", Config{}),
		*GetSimpleUser("find", Config{}),
	}

	if err := DB.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create users: %v", err)
	}

	t.Run("QParam and QList at the same time", func(t *testing.T) {
		req := &TestQParamReq{
			"find",
		}
		res := []model.SimpleUser{}
		var total int64 = -1
		if err := DB.Debug().Model(&model.SimpleUser{}).QParam(req).QList(req, &res, &total).Err(); err != nil {
			t.Errorf("errors happened when query first: %v", err)
		} else {
			CheckSimpleUser(t, res[0], users[0])
			CheckSimpleUser(t, res[1], users[2])
		}
	})
}
