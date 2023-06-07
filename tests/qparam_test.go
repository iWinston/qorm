package tests

//
//type TestQParamReq struct {
//	Name string `where:"=;Name"`
//}
//
//func TestQParam(t *testing.T) {
//	user1 := model.SimpleUser{}
//	user2 := model.SimpleUser{}
//	user3 := model.SimpleUser{}
//
//	GetUser("find", Config{}, &user1)
//	GetUser("find2", Config{}, &user2)
//	GetUser("find", Config{}, &user3)
//	users := []model.SimpleUser{
//		user1,
//		user2,
//		user3,
//	}
//
//	if err := DB.Create(&users).Error; err != nil {
//		t.Fatalf("errors happened when create users: %v", err)
//	}
//
//	t.Run("QParam and QList at the same time", func(t *testing.T) {
//		req := &TestQParamReq{
//			"find",
//		}
//		res := []model.SimpleUser{}
//		var total int64 = -1
//		if err := DB.Debug().Model(&model.SimpleUser{}).QParam(req).QList(req, &res, &total).Err(); err != nil {
//			t.Errorf("errors happened when query first: %v", err)
//		} else {
//			CheckSimpleUser(t, res[0], users[0])
//			CheckSimpleUser(t, res[1], users[2])
//		}
//	})
//}
