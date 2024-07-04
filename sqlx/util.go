package sqlx

import (
	"reflect"
	"strings"
)

/*
获取模型的select字段

	type AA struct {
		ID    uint      `db:"id"`
		Ctime time.Time `db:"ctime"`
	}

var a = GetModelSelectField(AA{}) // return "id, ctime"
*/
func GetModelSelectField(model interface{}) string {
	return GetModelSelectFieldByTagName(model, "db")
}

/*
获取模型的select字段

	type AA struct {
		ID    uint      `db:"id"`
		Ctime time.Time `db:"ctime"`
	}

var a = GetModelSelectFieldByTagName(AA{}, "db") // return "id, ctime"
*/
func GetModelSelectFieldByTagName(model interface{}, tagName string) string {
	var selectAllFields []string
	rt := reflect.TypeOf(model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	for _, field := range reflect.VisibleFields(rt) {
		// 拿到所有db字段
		if field.Tag.Get(tagName) != "" {
			selectAllFields = append(selectAllFields, field.Tag.Get(tagName))
		}
	}
	return strings.Join(selectAllFields, ", ")
}
