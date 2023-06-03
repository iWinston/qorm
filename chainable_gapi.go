package qorm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TODO 是否要给gorm的api也加上报错就不会继续执行的特性
// 重要：不要在Qorm中直接修改Statement！！！

// AddSelect 在当前Selects的基础上追加select
func (qdb *DB) AddSelect(name ...string) *DB {
	qtx := qdb.getInstance()
	selects := append(qtx.Statement.Selects, name...)
	return qtx.Select(selects)
}

// Session create new db session
func (qdb *DB) Session(config *gorm.Session) *DB {
	qtx := &DB{
		clone:        2,
		QErr:         qdb.QErr,
		DB:           qdb.DB.Session(config),
		TagMeta:      qdb.TagMeta,
		param:        qdb.param,
		paramTagMeta: qdb.paramTagMeta,
	}
	if config.NewDB {
		qtx.clone = 1
	}
	if config.Initialized {
		qtx = qtx.getInstance()
	}
	return qtx
}

// Model specify the model you would like to run db operations
//
//	// update all users's name to `hello`
//	db.Model(&User{}).Update("name", "hello")
//	// if user's primary key is non-blank, will use it as condition, then will only update the user's name to `hello`
//	db.Model(&user).Update("name", "hello")
func (qdb *DB) Model(value interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Model(value)
	return qtx
}

// Clauses Add clauses
func (qdb *DB) Clauses(conds ...clause.Expression) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Clauses(conds...)
	return qtx
}

// Table specify the table you would like to run db operations
func (qdb *DB) Table(name string, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Table(name, args...)
	return qtx
}

// Distinct specify distinct fields that you want querying
func (qdb *DB) Distinct(args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Distinct(args...)
	return qtx
}

// Select specify fields that you want when querying, creating, updating
func (qdb *DB) Select(query interface{}, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Select(query, args...)
	return qtx
}

// Omit specify fields that you want to ignore when creating, updating and querying
func (qdb *DB) Omit(columns ...string) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Omit(columns...)
	return qtx
}

// Where add conditions
func (qdb *DB) Where(query interface{}, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Where(query, args...)
	return qtx
}

// Not add NOT conditions
func (qdb *DB) Not(query interface{}, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Not(query, args...)
	return qtx
}

// IfWhere add condition when arg is not zero
func (qdb *DB) IfWhere(query interface{}, arg interface{}) *DB {
	qtx := qdb.getInstance()
	if arg == nil || IsZeroOfUnderlyingType(arg) {
		return qtx
	}
	if arg, ok := arg.([]uint); ok {
		if len(arg) == 0 {
			return qtx
		}
	}
	qtx.DB = qtx.DB.Where(query, arg)
	return qtx
}

// Or add OR conditions
func (qdb *DB) Or(query interface{}, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Or(query, args...)
	return qtx
}

// Joins specify Joins conditions
//
//	db.Joins("Account").Find(&user)
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
//	db.Joins("Account", DB.Select("id").Where("user_id = users.id AND name = ?", "someName").Model(&Account{}))
func (qdb *DB) Joins(query string, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Joins(query, args...)
	return qtx
}

// Group specify the group method on the find
func (qdb *DB) Group(name string) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Group(name)
	return qtx
}

// Having specify HAVING conditions for GROUP BY
func (qdb *DB) Having(query interface{}, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Having(query, args...)
	return qtx
}

// Order specify order when retrieve records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func (qdb *DB) Order(value interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Order(value)
	return qtx
}

// Limit specify the number of records to be retrieved
func (qdb *DB) Limit(limit int) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Limit(limit)
	return qtx
}

// Offset specify the number of records to skip before starting to return the records
func (qdb *DB) Offset(offset int) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Offset(offset)
	return qtx
}

// Scopes pass current database connection to arguments `func(DB) DB`, which could be used to add conditions dynamically
//
//	func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
//	    return db.Where("amount > ?", 1000)
//	}
//
//	func OrderStatus(status []string) func (db *gorm.DB) *gorm.DB {
//	    return func (db *gorm.DB) *gorm.DB {
//	        return db.Scopes(AmountGreaterThan1000).Where("status in (?)", status)
//	    }
//	}
//
//	db.Scopes(AmountGreaterThan1000, OrderStatus([]string{"pay", "shipped"})).Find(&orders)
func (qdb *DB) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Scopes(funcs...)
	return qtx
}

// Preload preload associations with given conditions
//
//	db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
func (qdb *DB) Preload(query string, args ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Preload(query, args...)
	return qtx
}

func (qdb *DB) Attrs(attrs ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Attrs(attrs...)
	return qtx
}
func (qdb *DB) Assign(attrs ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Assign(attrs...)
	return qtx
}

func (qdb *DB) Unscoped() *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Unscoped()
	return qtx
}

func (qdb *DB) Raw(sql string, values ...interface{}) *DB {
	qtx := qdb.getInstance()
	qtx.DB = qtx.DB.Raw(sql, values...)
	return qtx
}
