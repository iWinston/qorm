package qorm

import (
	"context"
	"errors"
	"reflect"
	"strings"

	errors2 "github.com/pkg/errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
	TagMeta
	QErr error
	// 暂存param的TagMeta，主要起到性能优化的作用
	paramTagMeta TagMeta
	param        interface{}
	clone        int
}

func (qdb *DB) Panic() {
	err := qdb.Err()
	if err != nil {
		panic(err)
	}
}

func (qdb *DB) Err() error {
	err := qdb.DB.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors2.Wrap(err, "查找不到对应数据")
		}
		if mysqlError, ok := qdb.DB.Error.(*mysql.MySQLError); ok {
			switch mysqlError.Number {
			case 1062:
				//m := regexp.MustCompile(`Duplicate entry '(.*)' for.*`).FindStringSubmatch(mysqlError.Message)
				return errors2.Wrap(err, "重复提交")
			case 1054:
				return errors2.Wrap(err, "查找了不存在的字段名")
			case 1451:
				return errors2.Wrap(err, "该数据存在关联数据，不允许操作")
			case 1452:
				return errors2.Wrap(err, "该关联数据不存在")
			}
		}
		return errors2.Wrap(qdb.DB.Error, "Gorm报错")
	}
	return qdb.QErr
}

func New(db *gorm.DB) *DB {
	if _, ok := db.Plugins[gormPluginJoin]; !ok {
		if err := db.Use(&JoinBeforeFinisher{}); err != nil {
			db.Error = err
		}
	}
	if _, ok := db.Plugins[gormPluginWhere]; !ok {
		if err := db.Use(&WhereBeforeFinisher{}); err != nil {
			db.Error = err
		}
	}
	return &DB{
		DB:           db,
		clone:        1,
		TagMeta:      NewTagMeta(),
		paramTagMeta: NewTagMeta(),
	}
}

func (qdb *DB) meta(param interface{}) *DB {
	// 清空旧有的TagMeta
	qdb.paramTagMeta = NewTagMeta()
	var (
		paramType  = reflect.TypeOf(param) //通过反射获取type定义
		paramValue = reflect.ValueOf(param)
	)
	qdb.paramTagMeta.meta(paramType, paramValue)
	qdb.TagMeta.Preloads = append(qdb.TagMeta.Preloads, qdb.paramTagMeta.Preloads...)
	return qdb
}

// QParam = QWhere().QOrder()
func (qdb *DB) QParam(param interface{}) *DB {
	return qdb.QWhere(param).QOrder(param)
}

// QRes = QSelect().QOrder()
func (qdb *DB) QRes(res interface{}) *DB {
	return qdb.QSelect(res).QOrder(res)
}

// 实现链式调用
// 值得注意的是，这个函数只管理qorm.DB的会话，不涉及对gorm.DB的影响
// clone = 0 链式操作
// clone = 1 进行初始化
// clone = 2 新建会话并复制内容信息
func (qdb *DB) getInstance() *DB {
	if qdb.clone > 0 {
		// set clone to 0
		qtx := &DB{
			clone: 0,
			QErr:  qdb.QErr,
			DB:    qdb.DB,
		}

		if qdb.clone == 1 {
			// clone with new TagMeta and param
			qtx.TagMeta = NewTagMeta()
			qtx.paramTagMeta = NewTagMeta()
			qtx.Statement.Context = context.WithValue(qtx.Statement.Context, "TAG_META", qtx.paramTagMeta)
			qtx.param = nil
			// session 模式： 此模式会保留之前的statement
		} else {
			// with clone DB，clone TagMeta and clone param
			qtx.TagMeta = qdb.TagMeta
			qtx.param = qdb.param
			qtx.paramTagMeta = qdb.paramTagMeta
		}
		qtx.setTagMetaToCtx()
		return qtx
	}
	qdb.setTagMetaToCtx()
	return qdb
}

func (qdb *DB) initDB(param interface{}) *DB {
	qtx := qdb.getInstance()
	if param != nil && param != qtx.param {
		qtx.meta(param)
		qtx.param = param
	}
	return qtx
}

func (qdb *DB) setTagMetaToCtx() {
	qdb.DB.Statement.Context = context.WithValue(qdb.DB.Statement.Context, "TAG_META", qdb.TagMeta)
}

func (qdb *DB) getColumnName(name string) (columnName string) {
	var (
		tableName = GetTableName(qdb.DB)
		index     = strings.LastIndex(name, ".")
	)
	// 长度1代表是自定义字段名
	if index == -1 {
		columnName = QuoteTo(qdb.DB, tableName) + `.` + QuoteTo(qdb.DB, qdb.DB.NamingStrategy.ColumnName("", name))
	} else {
		columnName = QuoteTo(qdb.DB, name[0:index]) + `.` + QuoteTo(qdb.DB, qdb.DB.NamingStrategy.ColumnName("", name[index+1:]))
	}
	return
}

func (qdb *DB) assignModelValue(dest interface{}) {
	// assign model values
	if qdb.Statement.Model == nil {
		qdb.Statement.Model = dest
	} else if qdb.Statement.Dest == nil {
		dest = qdb.Statement.Model
	}
}

// WithContext change current instance db's context to ctx
func (qdb *DB) WithContext(ctx context.Context) *DB {
	return qdb.Session(&gorm.Session{Context: ctx})
}

// Debug start debug mode
func (qdb *DB) Debug() (tx *DB) {
	return qdb.Session(&gorm.Session{
		Logger: qdb.Logger.LogMode(logger.Info),
	})
}
