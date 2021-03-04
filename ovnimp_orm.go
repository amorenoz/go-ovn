package goovn

import (
	"reflect"
)

func (odbi *ovndb) setUUID(model Model, uuid string) {
	uField := reflect.Indirect(reflect.ValueOf(model)).FieldByName("UUID")
	uField.Set(reflect.ValueOf(uuid))
}

func (odbi *ovndb) getUUID(model Model) string {
	return reflect.Indirect(reflect.ValueOf(model)).FieldByName("UUID").Interface().(string)
}
