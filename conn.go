package qorm

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Open(dial gorm.Dialector, opts ...gorm.Option) (*DB, error) {
	db, err := gorm.Open(dial, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "打开数据库连接错误")
	}
	return New(db), err
}
