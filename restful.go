package qorm

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

// TODO 待验证
// 对于非Model的结构体，Gorm的智能筛选字段和Preload是不能在一次链式操作中完成的。
//具体原因参考以下链接：https://github.com/go-gorm/gorm/issues/4015
// Qorm的Take和Find对此做了特殊处理，Take或者Find单个结构体时会先preload，然后智能筛选字段（智能筛选字段会覆盖原来的preload）
//但是Find数组时，目前只能先智能筛选字段，然后preload（preload会覆盖原来的智能筛选字段） 对于这种设计上的不一致性，后续再进行处理，目前难以实现统一

// QTake 兼容了同时Select和Preload
func (qdb *DB) QTake(dest interface{}, conds ...interface{}) *DB {

	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.assignModelValue(dest)

	if len(qtx.Statement.Preloads) == 0 {
		qtx.Take(dest, conds...)
		return qtx
	} else {
		// 这里是为了防止select里不包含外键字段,所以设置QueryFields模式，会根据当前model的所有字段名称进行select
		selects := qtx.Statement.Selects
		takeTX := qtx.Session(&gorm.Session{QueryFields: true})
		takeTX.Statement.Selects = nil
		takeTX.Session(&gorm.Session{SkipHooks: true}).Take(qtx.Statement.Model, conds...)
		if err := takeTX.Err(); err != nil {
			return takeTX
		}

		if err := ToStruct(qtx.Statement.Model, dest); err != nil {
			qtx.QErr = errors.Wrap(err, "把preload的Model复制到结果时出错")
			return qtx
		}
		// TODO 判断是否有select别名字段或者外键字段，当有的时候才select
		qtx.Statement.Selects = selects
		qtx.Statement.Preloads = nil
		qtx.Take(dest, conds...)
	}
	return qtx
}

// QFind 兼容了同时Select和Preload
func (qdb *DB) QFind(dest interface{}, conds ...interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.assignModelValue(dest)
	if len(qtx.Statement.Preloads) == 0 {
		qtx.Find(dest, conds...)
		return qtx
	}

	destTyp := reflect.TypeOf(dest).Elem()
	if destTyp.Kind() == reflect.Struct {
		// 这里是为了防止select里不包含外键字段,所以设置QueryFields模式，会根据当前model的所有字段名称进行select
		selects := qtx.Statement.Selects
		findTX := qtx.Session(&gorm.Session{QueryFields: true, SkipHooks: true})
		findTX.Find(qtx.Statement.Model)
		findTX.Statement.Selects = nil
		if err := findTX.Err(); err != nil {
			return findTX
		}

		if err := ToStruct(qtx.Statement.Model, dest); err != nil {
			qtx.QErr = errors.Wrap(err, "把preload的Model复制到结果时出错")
			return qtx
		}

		// TODO 判断是否有select别名字段或者外键字段，当有的时候才select
		qtx.Statement.Selects = selects
		qtx.Statement.Preloads = nil
		qtx.Find(dest, conds...)
		return qtx
	}

	// TODO 判断是否有select别名字段或者外键字段，当有的时候才select
	preloads := qtx.Statement.Preloads
	qtx.Statement.Preloads = nil
	findTX := qtx.Session(&gorm.Session{})
	findTX.Find(dest, conds...)
	if err := findTX.Err(); err != nil {
		return findTX
	}

	// 这里是为了防止select里不包含外键字段,所以设置QueryFields模式，会根据当前model的所有字段名称进行select
	arrType := reflect.SliceOf(reflect.TypeOf(qtx.Statement.Model).Elem())
	arr := reflect.New(arrType).Interface()
	qtx.Statement.Preloads = preloads
	qtx.Statement.Selects = nil
	qtx.Session(&gorm.Session{QueryFields: true}).Find(arr, conds...)
	if err := qtx.Err(); err != nil {
		return qtx
	}

	if err := ToStruct(arr, dest); err != nil {
		qtx.QErr = errors.Wrap(err, "把preload的Model复制到结果时出错")
		return qtx
	}
	return qtx
}

// QCreateOne 将参数赋值给模型后创建
func (qdb *DB) QCreateOne(param interface{}) *DB {
	v := reflect.ValueOf(param)
	if v.IsNil() {
		qdb.QErr = gorm.ErrInvalidValue
		return qdb
	}
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.initDB(param)
	printFileWithLineNum(qtx)
	m := qtx.Statement.Model
	if err := ToStruct(param, m); err != nil {
		qtx.QErr = errors.Wrap(err, "将param复制到Model时出错")
		return qtx
	}
	return qtx.Create(m)
}

// QPatchOne 查找模型之后，将param赋值给模型，然后Updates, 不会更新零值，不会更新关联模型
func (qdb *DB) QPatchOne(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	m := qtx.Statement.Model
	takeTX := qtx.Session(&gorm.Session{}).Clauses(clause.Locking{Strength: "UPDATE"}).Take(m)
	if err := takeTX.Err(); err != nil {
		return takeTX
	}
	if err := ToStruct(param, m); err != nil {
		qtx.QErr = errors.Wrap(err, "将param复制到Model时出错")
		return qtx
	}
	return qtx.Session(&gorm.Session{FullSaveAssociations: true}).Updates(m)
}

// QPutOne 查找模型之后，将param赋值给模型，然后Save
func (qdb *DB) QPutOne(param interface{}) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	m := qtx.Statement.Model
	takeTX := qtx.Session(&gorm.Session{}).Clauses(clause.Locking{Strength: "UPDATE"}).Take(m)
	if err := takeTX.Err(); err != nil {
		return takeTX
	}
	if err := ToStruct(param, m); err != nil {
		qtx.QErr = errors.Wrap(err, "将param复制到Model时出错")
		return qtx
	}
	return qtx.Session(&gorm.Session{FullSaveAssociations: true}).Save(m)
}

// QDeleteOne 查找模型之后，如果存在的话，将其删除
func (qdb *DB) QDeleteOne() *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	m := qtx.Statement.Model
	takeTX := qdb.Session(&gorm.Session{}).Take(m)
	if err := takeTX.Err(); err != nil {
		return takeTX
	}
	return qtx.Delete(m)
}

// QList 返回分页列表
func (qdb *DB) QList(param interface{}, res interface{}, total *int64) *DB {
	if qdb.Err() != nil {
		return qdb
	}
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	countTX := qtx.Session(&gorm.Session{}).
		QWhere(param).
		Count(total)
	if err := countTX.Err(); err != nil {
		return countTX
	}
	return qtx.QWhere(param).QOrder(param).Scopes(Paginate(param)).QOrder(res).QSelect(res).QFind(res)
}

func Paginate(req interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var page page
		ToStruct(req, &page)
		if page.Current > 0 && page.PageSize > 0 {
			offset := (page.Current - 1) * page.PageSize
			return db.Offset(offset).Limit(page.PageSize)
		}
		return db
	}
}

func PaginateRaw(sql string, req interface{}) string {
	var page page
	ToStruct(req, &page)
	if page.Current > 0 && page.PageSize > 0 {
		offset := (page.Current - 1) * page.PageSize
		sql += fmt.Sprintf(" LIMIT %v OFFSET %v", page.PageSize, offset)
	}
	return sql

}

type page struct {
	Current  int
	PageSize int
}
