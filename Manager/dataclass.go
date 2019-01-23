package Data

import (
	"errors"
	"reflect"
)

type dataClass struct {
	device    string
	lastLogin string
	this *map[string]interface{}
}

func (obj dataClass)GetMap() map[string]interface{} {
	if obj.this == nil {
		var key= reflect.TypeOf(obj)
		var value= reflect.ValueOf(obj)
		obj.this = new(map[string]interface{})
		for i := 0; i < key.NumField(); i++ {
			if key.Field(i).Name != "this" {
				(*obj.this)[key.Field(i).Name] = value.Field(i)
			}
		}
	}
	return *obj.this
}

func (obj dataClass)Set(key string,value interface{}) error {
	if key!="this" {
		return errors.New("dataClass write access error")
	}
	var field =reflect.ValueOf(&obj).FieldByName(key)
	if !field.IsValid(){
		return errors.New("can't find target element")
	}
	if field.Type()==reflect.ValueOf(value).Type() {
		field.Set(reflect.ValueOf(value))
	}else{
		return errors.New("field value type error")
	}
	return nil
}

