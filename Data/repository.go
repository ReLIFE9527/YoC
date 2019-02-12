package Data

import (
	"reflect"
	"time"
)

type repository struct {
	id        string
	key       string
	lastLogin time.Time
	this      *map[string]interface{}
}

func (stock *repository) GetMap() map[string]interface{} {
	var key = reflect.TypeOf(stock)
	var value = reflect.ValueOf(stock)
	stock.this = new(map[string]interface{})
	for i := 0; i < key.NumField(); i++ {
		if key.Field(i).Name != "this" {
			(*stock.this)[key.Field(i).Name] = value.Field(i)
		}
	}
	return *stock.this
}
