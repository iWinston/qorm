package qorm

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

// WhereBeforeFinisher 在执行gorm 增删改查之前，先执行qorm的where
// gorm 插件文档 https://gorm.io/zh_CN/docs/write_plugins.html
type WhereBeforeFinisher struct{}

func (p *WhereBeforeFinisher) Name() string {
	return gormPluginWhere
}

func (p *WhereBeforeFinisher) Initialize(db *gorm.DB) error {
	if err := db.Callback().Query().Before(gormPluginJoin).Register(gormPluginWhere, whereCallBack); err != nil {
		return errors.New("注册Query Where插件失败")
	}
	if err := db.Callback().Update().Before(gormPluginJoin).Register(gormPluginWhere, whereCallBack); err != nil {
		return errors.New("注册Update Where插件失败")
	}
	if err := db.Callback().Delete().Before(gormPluginJoin).Register(gormPluginWhere, whereCallBack); err != nil {
		return errors.New("注册Delete Where插件失败")
	}
	return nil
}

func whereCallBack(db *gorm.DB) {
	tagMetaVar := db.Statement.Context.Value("TAG_META")
	if tagMetaVar == nil {
		return
	}
	tagMeta := tagMetaVar.(TagMeta)
	for _, where := range tagMeta.Wheres {
		columnName := getColumnName(db, where.Name)
		genCondition(db, columnName, where.Flag, where.Value)
	}
	tagMeta.Wheres = []meta{}
	db.Statement.Context = context.WithValue(db.Statement.Context, "TAG_META", tagMeta)
}

func getColumnName(db *gorm.DB, name string) (columnName string) {
	var (
		tableName = db.Statement.Schema.Table
		arr       = strings.Split(name, ".")
	)
	// 长度1代表是自定义字段名
	if len(arr) == 1 {
		columnName = QuoteTo(db, tableName) + `.` + QuoteTo(db, db.NamingStrategy.ColumnName("", arr[0]))
	} else {
		columnName = QuoteTo(db, arr[len(arr)-2]) + `.` + QuoteTo(db, db.NamingStrategy.ColumnName("", arr[len(arr)-1]))
	}
	return
}

func genCondition(db *gorm.DB, name, operator string, itemValue interface{}) {
	switch operator {
	case "_":
	case "!=":
		db.Where(name+" != ?", itemValue)
	case "in":
		db.Where(name+" in ?", itemValue)
	case "=":
		db.Where(name+" = ?", itemValue)
	case ">":
		db.Where(name+" > ?", itemValue)
	case "<":
		db.Where(name+" < ?", itemValue)
	case ">=":
		db.Where(name+" >= ?", itemValue)
	case "<=":
		db.Where(name+" <= ?", itemValue)
	case "like":
		db.Where(name+" LIKE ?", "%"+*itemValue.(*string)+"%")
	case "null":
		db.Where(name + " IS NULL")
	case "not null":
		db.Where(name + " IS NOT NULL")
	default:
		// TODO   支持更加sql化的语法 # - 当前字段 ? 字段值
		db.Where(operator, itemValue)
	}
}
