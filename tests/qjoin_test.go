package tests

import (
	"testing"

	"github.com/iWinston/qorm/tests/model"
	"gorm.io/gorm"
)

func TestNestJoins(t *testing.T) {
	coupon := &model.Coupon{
		AppliesToProducts: []*model.CouponProduct{
			{ProductId: "full-save-association-product1"},
		},
		AmountOff:  10,
		PercentOff: 0.0,
	}

	err := DB.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Create(coupon).Error
	if err != nil {
		t.Errorf("Failed, got error: %v", err)
	}

	if DB.First(&model.Coupon{}, "id = ?", coupon.ID).Error != nil {
		t.Errorf("Failed to query saved coupon")
	}

	if DB.First(&model.CouponProduct{}, "coupon_id = ? AND product_id = ?", coupon.ID, "full-save-association-product1").Error != nil {
		t.Errorf("Failed to query saved association")
	}

	orders := []model.Order{{Num: "order1", Coupon: coupon}, {Num: "order2", Coupon: coupon}}
	if err := DB.Create(&orders).Error; err != nil {
		t.Errorf("failed to create orders, got %v", err)
	}

	order := model.Order{}
	if err := DB.Model(&order).Debug().QJoins("Coupon.AppliesToProducts").First(&order).Err(); err != nil {
		t.Fatalf("Failed to load with joins, got error: %v", err)
	}
}

func TestJoins(t *testing.T) {
	user := *GetSimpleUser("joins-1", Config{Company: true, Manager: true, Account: true, Pets: 2})

	DB.Create(&user)

	var user2 model.SimpleUser
	if err := DB.Model(&user2).Debug().QJoins("Languages").First(&user2, "users.name = ?", user.Name).Err(); err != nil {
		t.Fatalf("Failed to load with joins, got error: %v", err)
	}

	//CheckUser(t, user, user2)
}
