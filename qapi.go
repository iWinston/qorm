package qorm

import (
	"reflect"

	"github.com/pkg/errors"

	"gorm.io/gorm"
)

// TenantModel is model with the tenant scope
func (qdb *DB) TenantModel(value interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Model(value)
	return qtx
}

func (qdb *DB) QOrder(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.initDB(param)
	qtx.TagMeta.Orders = append(qtx.TagMeta.Orders, qtx.paramTagMeta.Orders...)
	for _, order := range qtx.TagMeta.Orders {
		if relation := getRelation(order.Name); relation != "" {
			qtx.QJoins(relation)
		}
		columnName := qtx.getColumnName(order.Name)
		qtx.Order(columnName + " " + order.Flag)
	}
	return qtx
}

func (qdb *DB) QWhere(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.initDB(param)
	qtx.TagMeta.Wheres = append(qtx.TagMeta.Wheres, qtx.paramTagMeta.Wheres...)
	for _, where := range qtx.paramTagMeta.Wheres {
		if relation := getRelation(where.Name); relation != "" {
			qtx.QJoins(relation)
		}
	}
	return qtx
}

func (qdb *DB) QSelect(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.initDB(param)
	qtx.TagMeta.Selects = append(qtx.TagMeta.Selects, qtx.paramTagMeta.Selects...)
	for _, preload := range qtx.TagMeta.Preloads {
		isPreloaded := false
		for key, _ := range qtx.DB.Statement.Preloads {
			if key == preload {
				isPreloaded = true
			}
		}
		if !isPreloaded {
			qtx.Preload(preload)
		}
	}

	var selects []string
	for _, item := range qtx.TagMeta.Selects {
		if relation := getRelation(item.Name); relation != "" {
			qtx.QJoins(relation)
		}
		selects = append(selects, qtx.getColumnName(item.Name)+" as "+QuoteTo(qtx.DB, item.Flag))
	}
	return qtx.Select(selects)
}

// QAssociation 处理关联模型
func (qdb *DB) QAssociation(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.initDB(param)
	printFileWithLineNum(qtx)
	qtx.Associations = append(qtx.Associations, qtx.paramTagMeta.Associations...)
	if len(qtx.Associations) > 0 {
		modelType := reflect.TypeOf(qtx.Statement.Model).Elem()
		for _, association := range qtx.Associations {
			m := qtx.Statement.Model
			// 空数组直接清空关系
			if len(association.Value.([]uint)) == 0 {
				err := qtx.Session(&gorm.Session{NewDB: true}).Model(m).Association(association.Name).Clear()
				if err != nil {
					qtx.QErr = errors.Wrap(err, "清空关系失败")
				}
				continue
			} else {
				relationField, ok := modelType.FieldByName(association.Name)
				if !ok {
					qtx.QErr = errors.Errorf("%s 中没有 %s 关联字段", modelType, association.Name)
					continue
				}
				relationModel := reflect.New(relationField.Type).Interface()
				if err := qtx.Session(&gorm.Session{NewDB: true, SkipHooks: true, Initialized: true}).DB.Model(relationModel).Find(relationModel, association.Value).Error; err != nil {
					qtx.QErr = errors.Wrap(err, "找不到对应id的值")
					continue
				}
				err := qtx.Session(&gorm.Session{FullSaveAssociations: true, NewDB: true, SkipHooks: true, Initialized: true}).Model(m).Association(association.Name).Replace(relationModel)
				if err != nil {
					qtx.QErr = errors.Wrap(err, "更新关系失败")
				}
			}
		}
	}
	return qtx
}

// QJoins not only support one to one and many to one, but also support one to many and many to many
//
//	db.Joins("Account").Find(&user)
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
func (qdb *DB) QJoins(query string, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.TagMeta.Joins = append(qtx.TagMeta.Joins, joinMeta{
		Query: query,
		Args:  args,
	})
	return qtx
}

// QTable 可以通过传入一个函数的方式动态修改表名
func (qdb *DB) QTable(f func(*gorm.DB) (string, error)) *DB {
	qtx := qdb.getInstance()
	if tableName, err := f(qtx.DB); err != nil {
		qtx.QErr = err
	} else {
		qtx.Table(tableName)
	}
	return qtx
}

// QScope 可以通过传入一个函数的方式动态添加语句
func (qdb *DB) QScope(f func(*DB) *DB) *DB {
	qtx := qdb.getInstance()
	qtx = f(qtx)
	return qtx
}
