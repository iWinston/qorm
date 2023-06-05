package qorm

import (
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func QuoteTo(db *gorm.DB, str string) string {
	builder := strings.Builder{}
	db.Dialector.QuoteTo(&builder, str)
	return builder.String()
}

func GetTableName(tx *gorm.DB) (tableName string) {
	tableName = tx.Statement.Table
	if tableName == "" {
		tableName = tx.NamingStrategy.TableName(reflect.TypeOf(tx.Statement.Model).Elem().Name())
	}
	if tabler, ok := tx.Statement.Model.(schema.Tabler); ok {
		tableName = tabler.TableName()
	}
	return
}

func getRelation(name string) (relation string) {
	index := strings.LastIndex(name, ".")
	if index != -1 {
		relation = name[0:index]
	}
	return
}

func getInfoFromTag(tag, defaultFlag, defaultName string) (flag, name, relation string) {
	flag = defaultFlag
	name = defaultName
	if tag != "" {
		arr := strings.Split(tag, ";")
		if len(arr) == 1 {
			flag = arr[0]
		}
		if len(arr) == 2 {
			flag = arr[0]
			name = arr[1]

			// 别名有可能是联表字段
			relation = getRelation(name)
		}
	}
	return
}

func getFlagAndNameFromTag(tag, defaultFlag, defaultName string) (flag, name string) {
	flag = defaultFlag
	name = defaultName
	if tag != "" {
		arr := strings.Split(tag, ";")
		if len(arr) == 1 {
			flag = arr[0]
		}
		if len(arr) == 2 {
			flag = arr[0]
			name = arr[1]
		}
	}
	return
}

var qormSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	qormSourceDir = regexp.MustCompile(`util\.go`).ReplaceAllString(file, "")
}

func fileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, qormSourceDir) || strings.HasSuffix(file, "_test.go")) {
			var funcName string
			if p, _, _, ok := runtime.Caller(i - 1); ok {
				arr := strings.Split(runtime.FuncForPC(p).Name(), ".")
				funcName = arr[len(arr)-1]
			}
			return funcName + " " + file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func printFileWithLineNum(db *DB) {
	db.Config.Logger.Info(db.Statement.Context, "qorm."+fileWithLineNum())
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
func GetDeepType(typ reflect.Type) reflect.Type {
	resKind := typ.Kind()
	if resKind == reflect.Array || resKind == reflect.Slice || resKind == reflect.Ptr {
		return GetDeepType(typ.Elem())
	} else {
		return typ
	}
}
func GetDeepValue(value reflect.Value) reflect.Value {
	resKind := value.Kind()
	if resKind == reflect.Ptr {
		return GetDeepValue(value.Elem())
	} else {
		return value
	}
}
func ToStruct(param, pointer interface{}) error {
	bytes, err := jsoniter.Marshal(&param)
	if err != nil {
		return err
	}
	if err = jsoniter.Unmarshal(bytes, pointer); err != nil {
		return err
	}
	return nil
}
