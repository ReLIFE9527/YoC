package Data

import (
	"reflect"
	"time"
)

type repository struct {
	Id        string
	Key       string
	LastLogin time.Time
	this      *map[string]interface{}
}

func (obj *repository) GetMap() map[string]interface{} {
	var key = reflect.TypeOf(obj)
	var value = reflect.ValueOf(obj)
	obj.this = new(map[string]interface{})
	for i := 0; i < key.NumField(); i++ {
		if key.Field(i).Name != "this" {
			(*obj.this)[key.Field(i).Name] = value.Field(i)
		}
	}
	return *obj.this
}
