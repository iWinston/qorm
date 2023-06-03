package qorm

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type TagMeta struct {
	Joins        []joinMeta
	Preloads     []string
	Selects      []meta
	Orders       []meta
	Wheres       []meta
	Associations []meta
}

type joinMeta struct {
	Query string
	Args  []interface{}
	Alias string
}

type meta struct {
	Name  string
	Value interface{}
	Flag  string
}

func NewTagMeta() TagMeta {
	return TagMeta{
		[]joinMeta{},
		[]string{},
		[]meta{},
		[]meta{},
		[]meta{},
		[]meta{},
	}
}

func (m *TagMeta) meta(typ reflect.Type, value reflect.Value) {
	typ = GetDeepType(typ)
	value = GetDeepValue(value)

	for i := 0; i < typ.NumField(); i++ {
		itemType := typ.Field(i)
		// 默认值是零值
		itemValue := reflect.New(itemType.Type)
		// 非数组则设为结构体的数值 TODO QParam如果传一个数组如何处理，这里是处理为零值
		if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
			itemValue = value.Field(i)
		}
		selectTag, isSelectTagExisted := itemType.Tag.Lookup("select")
		// 是匿名结构体，递归字段
		if itemType.Anonymous && selectTag != "_" {
			m.meta(itemType.Type, itemValue)
			continue
		}

		m.setOrders(itemType)
		m.setWheres(itemType, itemValue)
		m.setAssociations(itemType, itemValue)
		isPreloadTagExisted := m.setPreloads(itemType)
		tagV := itemType.Tag.Get("gorm")

		// 当前字段的类型的名称包含model则不应该被select，如果有preload或者foreignKey，也不应该被select
		shouldNotSelect := strings.Contains(itemType.Type.String(), "model.") || strings.Contains(tagV, "foreignKey") || isPreloadTagExisted
		if isSelectTagExisted || !shouldNotSelect {
			m.setSelects(itemType)
		}

	}
}

func (m *TagMeta) setOrders(typ reflect.StructField) {
	var tag, isTagExisted = typ.Tag.Lookup("order")

	if isTagExisted {
		var sequence, itemName = getFlagAndNameFromTag(tag, "desc", typ.Name)
		m.Orders = append(m.Orders, meta{
			Name: itemName,
			Flag: sequence,
		})
	}
}

func (m *TagMeta) setWheres(typ reflect.StructField, value reflect.Value) {
	var tag, isTagExisted = typ.Tag.Lookup("where")

	if isTagExisted {
		var operation, itemName = getFlagAndNameFromTag(tag, "=", typ.Name)
		// 此处是默认所有的Dto都是指针类型,结构体或者数组
		// 如果不是nullable的字段，则为空不传条件
		if IsZeroOfUnderlyingType(value.Interface()) {
			_, nullableExist := typ.Tag.Lookup("nullable")
			if !nullableExist {
				return
			}
		}
		m.Wheres = append(m.Wheres, meta{
			Name:  itemName,
			Flag:  operation,
			Value: value.Interface(),
		})
	}
}

func (m *TagMeta) setPreloads(typ reflect.StructField) bool {
	var preloadTag, isPreloadTagExisted = typ.Tag.Lookup("preload")
	if isPreloadTagExisted {
		if preloadTag == "" {
			preloadTag = typ.Name
		}
		m.Preloads = append(m.Preloads, preloadTag)
	}
	return isPreloadTagExisted
}

func (m *TagMeta) setSelects(typ reflect.StructField) {
	selectTag := typ.Tag.Get("select")
	if selectTag != "_" {
		if selectTag == "" {
			selectTag = typ.Name
		}
		m.Selects = append(m.Selects, meta{
			Name: selectTag,
			Flag: typ.Name,
		})
	}
}

func (m *TagMeta) setAssociations(typ reflect.StructField, value reflect.Value) {
	var tag, isTagExisted = typ.Tag.Lookup("association")
	if isTagExisted {
		// 不传的情况不处理
		if IsZeroOfUnderlyingType(value.Interface()) {
			return
		}
		if tag == "" {
			if strings.HasSuffix(typ.Name, "Ids") {
				tag = tag[:len(typ.Name)-3]
			} else {
				panic(errors.New("association Tag 为空，并且字段名不是以Ids结尾"))
			}
		}
		m.Associations = append(m.Associations, meta{
			Name:  tag,
			Value: value.Interface().([]uint),
		})
	}
}
