# QORM

## 定位

在GORM的基础上，增加对标签的支持，强化了指定结构体查询字段，智能筛选字段，多对多Join等功能。

## 设计理念

尽量保持和GORM的语法一致，避免过多的理解和学习成本。

以下摘录Gorm的部分文档，方便理解增强内容。

### GORM的指定结构体查询字段和智能筛选字段说明

#### 指定结构体查询字段

当使用 struct 进行查询时，你可以通过向 Where() 传入 struct 来指定查询条件的字段、值、表名，例如：

```
db.Where(&User{Name: "jinzhu"}, "name", "Age").Find(&users)
// SELECT * FROM users WHERE name = "jinzhu" AND age = 0;

db.Where(&User{Name: "jinzhu"}, "Age").Find(&users)
// SELECT * FROM users WHERE age = 0;
```

#### 智能选择字段

GORM 允许通过 Select 方法选择特定的字段，如果您在应用程序中经常使用此功能，你也可以定义一个较小的结构体，以实现调用 API 时自动选择特定的字段，例如：

```
type User struct {
ID     uint
Name   string
Age    int
Gender string
// 假设后面还有几百个字段...
}

type APIUser struct {
ID   uint
Name string
}

// 查询时会自动选择 `id`, `name` 字段
db.Model(&User{}).Limit(10).Find(&APIUser{})
// SELECT `id`, `name` FROM `users` LIMIT 10
```

## QORM对GORM的增强

### 建立数据库连接并返回DB对象
可以使用 qorm.Open(), 来建立连接返回DB

### QWhere指定结构体查询字段

Gorm的Where的局限在于：

1. 只支持相等条件，不支持大于小于等条件
2. 只支持模型里有的字段，不支持别名和联表

Qorm通过在QWhere(param)的参数上添加where标签来实现对以上两点的支持 具体标签规则通过以下例子来说明：

```
type CatParam struct {
    Name *string // 没加标签的，默认不添加为条件
    Color *string `where:""` // 空的where标签，默认添加=条件
    // SELECT * FROM cat WHERE color = "";
    Age *uint `where:">"` // where后面可以加个比较符号，比如>,<,=,like等，如此时Age的值是2，那么产生的sql是"Where age > 2"
    // SELECT * FROM cat WHERE age > 0;
    Sex *string `where:"=;Gender"` // where后面可以用分号割开，加第二个参数，自定义字段名，如此时Sex的值是1，那么产生的sql是"Where gender = 1"
    // SELECT * FROM cat WHERE gender = 0;
    ParentName *string `where:"=;Parent.Name"` // 如果自定义字段名是级联形式，会自动触发联表操作
    // SELECT * FROM cat LEFT JOIN Parent ON Parent.Id = cat.ParentId WHERE Parent.Name = "";
}
```

调用方式：

```
var param CatParam
db.Model(&model.Cat{}).QWhere(&param)
```

### QSelect指定结构体查询字段

Gorm的Select的局限在于

1. 不支持自动Join
2. 不支持自动Preload
3. 不支持AddSelect

Qorm通过在QSelect(param)的参数上添加select标签来实现对Join和Preload的支持：

标签规则：

```
type CatRes struct {
    Name *string // 没有添加标签的，默认select
    Age *uint `select:"_"` // 下划线代表跳过不select，这种情况下，该字段会是零值
    ParentName *string `select:"Parent.Name"` // 这种情况会触发联表操作，且会筛选对应的值
    Parent *Parent `preload:"Parent"` // 自动进行Preload操作，这里填入的参数要和模型里的关系对应，可以和字段名不同
    Parent *Parent `preload:""` //不写的话，默认参数是字段名Parent
}
```

调用方式：

```
var res CatRes
db.Model(&model.Cat{}).QSelect(&res).Take(&res)
```

一般情况下，我们不会用到QSelect()这个API，因为在没有特殊指定Select的情况下，可以直接用Take()或者, 内部会调用一次QSelect()

```
var res CatRes
db.Model(&model.Cat{}).Take(&res)
```

特殊备注：

```
对于非Model的结构体，Gorm的智能筛选字段和Preload是不能在一次链式操作中完成的。
具体原因参考以下链接：https://github.com/go-gorm/gorm/issues/4015
Qorm的Take和Find对此做了特殊处理，Take或者Find单个结构体时会先preload，然后智能筛选字段（智能筛选字段会覆盖原来的preload）
但是Find数组时，目前只能先智能筛选字段，然后preload（preload会覆盖原来的智能筛选字段） 对于这种设计上的不一致性，后续再进行处理，目前难以实现统一
```

此外，Qorm新增了AddSelect方法，语法与Select相同，会在原来的Select上继续新增筛选项

### QOrder:支持根据结构体字段排序

QOrder的标签支持方式和QWhere是类似的

```
type CatParam struct {
    Name *string // 没加标签的，默认不添加排序
    Color *string `order:""` // 空的order标签，默认添加DESC条件
    // SELECT * FROM cat ORDERBY color DESC;
    Age *uint `order:"DESC"` // order后面可以加个排序方式，DESC或者ASC
    // SELECT * FROM cat ORDERBY age DESC;
    Sex *string `order:"DESC;Gender"` // where后面可以用分号割开，加第二个参数，自定义字段名，如此时Sex的值是1，那么产生的sql是"Where gender = 1"
    // SELECT * FROM cat OREDER gender DESC;
    ParentName *string `where:"DESC;Parent.Name"` // 如果自定义字段名是级联形式，会触发联表操作
    // SELECT * FROM cat LEFT JOIN Parent ON Parent.Id = cat.ParentId ORDER Parent.Name DESC;
}
```

调用方式：

```
var param CatParam
db.Model(&model.Cat{}).QOrder(&param)
```

### QAssociation：支持关联模型操作

Gorm 支持关联模式( https://gorm.io/zh_CN/docs/associations.html )，部分语法如下：

清空关联，删除源模型与关联之间的所有引用，但不会删除这些关联

```
db.Model(&user).Association("Languages").Clear()
```

替换关联, 用一个新的关联替换当前的关联

```
db.Model(&user).Association("Languages").Replace([]Language{languageZH, languageEN})

db.Model(&user).Association("Languages").Replace(Language{Name: "DE"}, languageEN)
```

这个语法要求先Find出对应的关联模型，在上例中是Language{}。 在Qorm中，我们支持传入Id数组，Qorm自动执行先Find后Replace的操作。 如果传入的是空数组，则执行Clear操作。

规则如下：

```
type CatParam struct {
    OwnerIds []uint `association:"Owners"`
    // db.Where("in IN ?"),Find(&owners)
    // db.Model(&cat).Association("Owners").Replace([]Owner{})
}
```

调用方式：

```
var param CatParam
db.Model(&model.Cat{}).QAssociation(&param)
```

### QJoins

Gorm的Joins函数只能针对1对1和多对1关系，不支持对多关系，QJoins则全面支持：

调用方式：

```
// QJoins("Owners")，参数是关联模型字段
db.Model(&model.Cat{}).QJoins("Owners")
```

## QORM封装的增删改查接口

### QCreateOne

QCreateOne 将参数赋值给模型后创建，不需要显式的指明 param 的表名是user表
```go
    type User struct{
        gorm.Model
        Name      string
        Age       uint
    }
	type SimpleUser struct {
		Name string
		Age uint
    }   
	
	DB.Model(&User{}).QCreateOne(&SimpleUser{
		Name:"zhangsan",
		Age: 18
})
// INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`age`) VALUES ('2023-06-11 20:35:12.237','2023-06-11 20:35:12.237',NULL,'create',18)
```

### QPatchOne

QPatchOne 查找模型之后，将param赋值给模型，然后Update

### QDeleteOne

QDeleteOne 查找模型之后，如果存在的话，将其删除，不存在的话，会抛出gorm.ErrRecordNotFound
```go
    type User struct{
        gorm.Model
        Name      string
        Age       uint
    }
	type SimpleUser struct {
		Id  uint `where:""`
		Age uint `where:">="`
    }   
	
	DB.Model(&User{}).QWhere(&SimpleUser{Id:1,Age:18}).QDeleteOne()
// SELECT * FROM `users` WHERE `users`.`id` = 2 AND `users`.`age` >= 18 AND `users`.`deleted_at` IS NULL LIMIT 1
// UPDATE `users` SET `deleted_at`='2023-06-11 20:38:08.247' WHERE `users`.`id` = 2 AND `users`.`age` >= 18 AND `users`.`id` = 2 AND `users`.`deleted_at` IS NULL


```


### QList

QList 返回分页列表，语法糖，无须再添加QWhere等条件

## 错误处理

与Gorm不同，Qorm的错误获取，调用的是Err()方法；使用方式如下：

```
cat := &model.Cat{}
if err := db.Find(cat).Err(); err != nil {
    return err
}
```

Err() 会优先展示GormAPI产生的错误，再展示QormAPI产生的错误

## 语法规则

在此说明下Qorm的语法设计规则：

1. 基于struct的Tag，标签名有：where，order，select，preload，association
2. 标签内容的语法格式"字段1;字段2"
3. 字段1一般用来作为标签的主要参数，在where中是条件，在order中是排序方式，在select中的筛选字段，在preload和association是关联关系
4. 字段2一般用来做别名，目前只在where和order中使用
5. 在标签内容中的字段名要和模型的字段名对应，一般都是大写，如Parent.Name

## API汇总

API命名规则，如果增强Gorm，并且和Gorm兼容的，则不加Q，如果是与Gorm无关的，则加Q

1. 和Meta相关的API

- QWhere 生成where语句
- QOrder 生成Order语句
- QSelect
- QAssociation
- QParam 语法糖 = QWhere().QOrder
- join和preload按需自动执行

原理说明：Qorm会分析参数的标签，然后将信息储存到db.Meta中，调用QWhere()等API时，会拿出db.Meta中的对应信息生成sql。

以上API都依赖于Model,所以使用前必须使用调用Model函数！

2. 和Restful相关的API

- QTake 对Take进行封装
- QFind 对Find进行封装
- QCreateOne 将单个param赋值给模型,然后Create。
- QPatchOne 查找模型之后，将param赋值给模型，然后Update
- QPutOne 查找模型之后,将param赋值给模型,然后Save
- QDeleteOne 根据where条件查找模型之后，如果数据存在的，会将其删除,不存在会抛出gorm.ErrNotFoundRecord。
- QList 返回分页列表

3. 其他注意事项

- 使用子查询时,因为qorm本身并不是ORM,因此需要注意传入的子查询语句,应当是gorm.DB,而不是qorm.DB,eg: subQuery= qormDB.model(&user{}).DB。一个例子可参见 qcreate_one_test.TestCreateFromSubQuery()