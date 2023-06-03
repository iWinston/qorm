package qorm

import (
	"database/sql"

	"gorm.io/gorm"
)

// Create insert the value into database
// will call gorm.DB.Create and set operation log value to context
func (qdb *DB) Create(value interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Create(value)
	return qtx
}

// CreateInBatches insert the value in batches into database
func (qdb *DB) CreateInBatches(value interface{}, batchSize int) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.CreateInBatches(value, batchSize)
	return qtx
}

// Save update value in database, if the value doesn't have primary key, will insert it
func (qdb *DB) Save(value interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Save(value)
	return qtx
}

// First find first record that match given conditions, order by primary key
func (qdb *DB) First(dest interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.First(dest, conds...)
	return qtx
}

// Take return a record that match given conditions, the order will depend on the database implementation
//func (qdb *DB) Take(dest interface{}, conds ...interface{}) *DB {
//	qtx := qdb.getInstance()
//	qtx.DB = qtx.DB.Take(dest, conds...)
//	return qtx
//}

// Last find last record that match given conditions, order by primary key
func (qdb *DB) Last(dest interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Last(dest, conds...)
	return qtx
}

// Take take records that match given conditions
func (qdb *DB) Take(value interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.assignModelValue(value)
	if qtx.Statement.Selects == nil {
		qtx.QSelect(value)
	}
	qtx.DB = qtx.DB.Take(value, conds...)
	return qtx
}

// Find  find records that match given conditions
func (qdb *DB) Find(value interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.assignModelValue(value)
	//TODO 待删除, 有旧项目在直接Find,因此保留
	if qtx.Statement.Selects == nil {
		qtx.QSelect(value)
	}
	qtx.DB.Find(value, conds...)
	return qtx
}

// Find find records that match given conditions
//func (qdb *DB) Find(dest interface{}, conds ...interface{}) *DB {
//	qtx := qdb.getInstance()
//	qtx.DB = qtx.DB.Find(dest, conds...)
//	return qtx
//}

// FindInBatches find records in batches
func (qdb *DB) FindInBatches(dest interface{}, batchSize int, fc func(tx *gorm.DB, batch int) error) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.FindInBatches(dest, batchSize, fc)
	return qtx
}

func (qdb *DB) FirstOrInit(dest interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.FirstOrInit(dest, conds...)
	return qtx
}

func (qdb *DB) FirstOrCreate(dest interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.FirstOrCreate(dest, conds...)
	return qtx
}

// Update update attributes with callbacks, refer: https://gorm.io/docs/update.html#Update-Changed-Fields
func (qdb *DB) Update(column string, value interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Update(column, value)
	return qtx
}

// Updates update attributes with callbacks, refer: https://gorm.io/docs/update.html#Update-Changed-Fields
func (qdb *DB) Updates(values interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.Session(&gorm.Session{}).DB.Updates(values)
	return qtx
}

func (qdb *DB) UpdateColumn(column string, value interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.UpdateColumn(column, value)
	return qtx
}

func (qdb *DB) UpdateColumns(values interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.UpdateColumns(values)
	return qtx
}

// Delete delete value match given conditions, if the value has primary key, then will including the primary key as condition
func (qdb *DB) Delete(value interface{}, conds ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Delete(value, conds...)
	return qtx
}

func (qdb *DB) Count(count *int64) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Count(count)
	return qtx
}

func (qdb *DB) Row() *sql.Row {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	return qtx.DB.Row()
}

func (qdb *DB) Rows() (*sql.Rows, error) {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	return qtx.DB.Rows()
}

// Scan scan value to a struct
func (qdb *DB) Scan(dest interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Scan(dest)
	return qtx
}

// Pluck used to query single column from a model as a map
//
//	var ages []int64
//	db.Model(&users).Pluck("age", &ages)
func (qdb *DB) Pluck(column string, dest interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Pluck(column, dest)
	return qtx
}

func (qdb *DB) ScanRows(rows *sql.Rows, dest interface{}) error {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	return qtx.DB.ScanRows(rows, dest)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit.
func (qdb *DB) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) (err error) {
	qtx := qdb.getInstance()
	return qtx.DB.Transaction(fc, opts...)
}

// Begin begins a transaction
func (qdb *DB) Begin(opts ...*sql.TxOptions) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Begin(opts...)
	return qtx
}

// Commit commit a transaction
func (qdb *DB) Commit() *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Commit()
	return qtx
}

// Rollback rollback a transaction
func (qdb *DB) Rollback() *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Rollback()
	return qtx
}

func (qdb *DB) SavePoint(name string) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.SavePoint(name)
	return qtx
}

func (qdb *DB) RollbackTo(name string) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.RollbackTo(name)
	return qtx
}

// Exec execute raw sql
func (qdb *DB) Exec(sql string, values ...interface{}) *DB {
	qtx := qdb.getInstance()
	printFileWithLineNum(qtx)
	qtx.DB = qtx.DB.Exec(sql, values...)
	return qtx
}
