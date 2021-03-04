package goovn

import (
	"fmt"
	"reflect"
)

type TableName = string

type Model interface {
	Table() TableName
}

type BaseModel struct {
	UUID string `ovs:"_uuid"` // Still not sure if ovs tag is needed
}

type DBModel struct {
	types map[TableName]reflect.Type
}

// NewModel returns a new instance of a model from a specific TableName
func (db DBModel) NewModel(table TableName) (Model, error) {
	mtype, ok := db.types[table]
	if !ok {
		return nil, fmt.Errorf("Table %s not found in Database Model", string(table))
	}
	model := reflect.New(mtype.Elem())
	return model.Interface().(Model), nil
}

func NewDBModel(models []Model) (*DBModel, error) {
	types := make(map[TableName]reflect.Type, len(models))
	for _, model := range models {
		if reflect.TypeOf(model).Kind() != reflect.Ptr {
			return nil, fmt.Errorf("Model is expected to be a pointer")
		}
		uField := reflect.Indirect(reflect.ValueOf(model)).FieldByName("UUID")
		if !uField.IsValid() || uField.Type().Kind() != reflect.String {
			return nil, fmt.Errorf("Model is expected to have a string field called UUID")

		}
		types[model.Table()] = reflect.TypeOf(model)
	}
	return &DBModel{
		types: types,
	}, nil
}
