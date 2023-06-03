package qorm

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JoinBeforeFinisher 在执行gorm 增删改查之前，先执行qorm的join
// gorm 插件文档 https://gorm.io/zh_CN/docs/write_plugins.html
type JoinBeforeFinisher struct{}

func (p *JoinBeforeFinisher) Name() string {
	return gormPluginJoin
}

func (p *JoinBeforeFinisher) Initialize(db *gorm.DB) error {
	if err := db.Callback().Query().Before(gormPluginQuery).Register(gormPluginJoin, joinCallBack); err != nil {
		return errors.New("注册Query Join插件失败")
	}

	if err := db.Callback().Update().Before(gormPluginUpdate).Register(gormPluginJoin, whereCallBack); err != nil {
		return errors.New("注册Update Join插件失败")
	}
	if err := db.Callback().Delete().Before(gormPluginDelete).Register(gormPluginJoin, whereCallBack); err != nil {
		return errors.New("注册Delete Join插件失败")
	}
	return nil
}

func joinCallBack(db *gorm.DB) {
	tagMetaVar := db.Statement.Context.Value("TAG_META")
	if tagMetaVar == nil {
		return
	}
	tagMeta := tagMetaVar.(TagMeta)
	for _, meta := range tagMeta.Joins {
		nestLevel := len(strings.Split(meta.Query, "."))
		switch true {
		case nestLevel == 1:
			{
				rel := db.Statement.Schema.Relationships.Relations[meta.Query]
				join(db, rel, meta)
			}
		case nestLevel <= 3:
			{
				nestJoin(db, meta)
			}
		default:
			{
				db.Error = errors.New("join的层级数量大于3，出于性能考虑，禁止join")
				return
			}
		}
	}
	tagMeta.Joins = []joinMeta{}
	db.Statement.Context = context.WithValue(db.Statement.Context, "TAG_META", tagMeta)
}

func nestJoin(db *gorm.DB, meta joinMeta) {
	joinArr := strings.Split(meta.Query, ".")
	// 第一级join从db.Statement.Schema，也就是model的关系开始
	rel := db.Statement.Schema.Relationships.Relations[joinArr[0]]
	join(db, rel, joinMeta{
		Query: joinArr[0],
		Args:  meta.Args,
	})
	// 后续的join从rel.FieldSchema，也就是关联的模型开始
	for i := 1; i < len(joinArr); i++ {
		rel = rel.FieldSchema.Relationships.Relations[joinArr[i]]
		join(db, rel, joinMeta{
			Query: joinArr[i],
			Args:  meta.Args,
			Alias: joinArr[i-1],
		})
	}

}

func join(db *gorm.DB, rel *schema.Relationship, meta joinMeta) {
	if rel == nil {
		db.Error = errors.Errorf("模型上找不到%s关系", meta.Query)
		return
	}

	joinName := QuoteTo(db, meta.Query)

	// TODO 感觉这种方式不是很优雅
	// 避免重复进行join
	isContains := false
	for _, join := range db.Statement.Joins {
		if join.Name == meta.Query ||
			strings.Contains(join.Name, joinName+" on") ||
			strings.Contains(join.Name, joinName+" ON") ||
			strings.Contains(join.Name, meta.Query+" on") ||
			strings.Contains(join.Name, meta.Query+" ON") {
			isContains = true
			break
		}
	}
	if isContains {
		return
	}

	tableName := QuoteTo(db, rel.Schema.Table)
	// 多次join的时候，要以上次的别名为准，否则会找不到字段
	if meta.Alias != "" {
		tableName = QuoteTo(db, meta.Alias)
	}
	relTableName := QuoteTo(db, rel.FieldSchema.Table)
	foreignKey := QuoteTo(db, rel.References[0].ForeignKey.DBName)
	primaryKey := QuoteTo(db, rel.References[0].PrimaryKey.DBName)
	switch rel.Type {
	case schema.HasOne:
		{
			db.Joins(fmt.Sprintf(`LEFT JOIN %s %s ON %s.%s = %s.%s`,
				relTableName, joinName, joinName, foreignKey, tableName, primaryKey), meta.Args...)
		}
	case schema.BelongsTo:
		{
			db.Joins(fmt.Sprintf(`LEFT JOIN %s %s ON %s.%s = %s.%s`,
				relTableName, joinName, joinName, primaryKey, tableName, foreignKey), meta.Args...)
		}
	case schema.HasMany:
		{
			db.Distinct().Joins(fmt.Sprintf(`LEFT JOIN %s %s ON %s.%s = %s.%s`,
				relTableName, joinName, joinName, foreignKey, tableName, primaryKey), meta.Args...)
		}
	case schema.Many2Many:
		{
			// 先join中间表
			midTableName := QuoteTo(db, rel.JoinTable.Table)
			midJoinName := QuoteTo(db, rel.JoinTable.Name)
			db.Distinct().Joins(fmt.Sprintf(`LEFT JOIN %s %s ON %s.%s = %s.%s`,
				midTableName, midJoinName, midJoinName, foreignKey, tableName, primaryKey), meta.Args...)

			// 再从中间表join真正要join的表
			join(db, rel.JoinTable.Relationships.Relations[rel.FieldSchema.Name], joinMeta{
				Query: rel.Name,
				Args:  meta.Args,
				Alias: midJoinName,
			})
		}
	default:
		{
			db.Error = errors.Errorf("不受支持的关系类型%s", rel.Type)
		}
	}

}
