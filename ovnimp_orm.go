package goovn

import (
	"fmt"
	"reflect"

	"github.com/ebay/libovsdb"
)

// GetByID is a generic Get function capable of returning (through a provided pointer)
// a instance of any row in the cache. It only works on ORM mode.
// 'result' must be a pointer to an Model that exists in the DBModel
// The main difference with Get() is that this function is O(1), while Get() is O(n)
func (odbi *ovndb) GetByID(result interface{}, uuid string) error {
	if odbi.mode != ORM {
		return fmt.Errorf("GetByID() is only available in ORM mode")
	}

	resultVal := reflect.ValueOf(result)
	if resultVal.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("GetByID() result must be a pointer")
	}

	table := odbi.findTable(resultVal.Type())
	if table == "" {
		return ErrorSchema
	}

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()
	tableCache, ok := odbi.ormCache[table]
	if !ok {
		return ErrorNotFound
	}

	elem, ok := tableCache[uuid]
	if !ok {
		return ErrorNotFound
	}

	resultVal.Elem().Set(reflect.Indirect(reflect.ValueOf(elem)))
	return nil
}

// Get is a generic Get function capable of returning (through a provided pointer)
// a instance of any row in the cache. It only works on ORM mode.
// 'result' must be a pointer to an Model that exists in the DBModel
//
// The way the cache is search depends on the fields already populated in 'result'
// Any table index (including _uuid) will be used for comparison. Additionally,
// any additional index provided as argument will be also used
func (odbi *ovndb) Get(model Model, index ...string) error {
	if odbi.mode != ORM {
		return fmt.Errorf("Get() is only available in ORM mode")
	}

	resultVal := reflect.ValueOf(model)
	if resultVal.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("Get() result must be a pointer")
	}

	table := odbi.findTable(resultVal.Type())
	if table == "" {
		return ErrorSchema
	}

	elem, err := odbi.ormFindInCache(table, model, index...)
	if err != nil {
		return err
	}
	resultVal.Elem().Set(reflect.Indirect(reflect.ValueOf(elem)))
	return nil

}

// List is a generic function capable of returning (through a provided pointer)
// a list of instances of any row in the cache. It only works on ORM mode.
// 'result' must be a pointer to an slice of the ORM structs that shall be retrived
// The items are appended to the given (pointer to) slice until its capability is reached.
// If the slice is null, all of the table cache will be returned
func (odbi *ovndb) List(result interface{}) error {
	if odbi.mode != ORM {
		return fmt.Errorf("List() is only available in ORM mode")
	}

	resultPtr := reflect.ValueOf(result)
	if resultPtr.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("List() result must be a pointer")
	}

	resultVal := reflect.Indirect(resultPtr)
	if resultVal.Type().Kind() != reflect.Slice {
		return fmt.Errorf("List() result must be a pointer to slice")
	}

	// DBModel stores pointer to structs, slice should have structs, so calling PtrTo
	table := odbi.findTable(reflect.PtrTo(resultVal.Type().Elem()))
	if table == "" {
		return fmt.Errorf("Schema error: finding table for types %s. Table content %+v", reflect.PtrTo(resultVal.Type().Elem()), odbi.dbModel.types)
	}

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	tableCache, ok := odbi.ormCache[table]
	if !ok {
		return ErrorNotFound
	}

	// If given a null slice, fill it in the cache table completely, if not, just up to
	// its capability
	if resultVal.IsNil() {
		resultVal.Set(reflect.MakeSlice(resultVal.Type(), 0, len(tableCache)))
	}
	i := resultVal.Len()
	for _, elem := range tableCache {
		if i >= resultVal.Cap() {
			break
		}
		resultVal.Set(reflect.Append(resultVal, reflect.Indirect(reflect.ValueOf(elem))))
		i++
	}
	return nil
}

// Create is a generic function capable of creating any row in the DB
// It only works on ORM mode.
// A valud Model (pointer to object) must be provided.
func (odbi *ovndb) Create(model Model) (*OvnCommand, error) {
	var uuid string
	var err error

	if odbi.mode != ORM {
		return nil, fmt.Errorf("Create() is only available in ORM mode")
	}

	table := odbi.findTable(reflect.ValueOf(model).Type())
	if table == "" {
		return nil, ErrorSchema
	}

	if muuid := odbi.getUUID(model); muuid != "" {
		uuid = muuid
	} else {
		uuid, err = newRowUUID()
		if err != nil {
			return nil, err
		}
	}

	// Check if exists already
	_, err = odbi.ormFindInCache(table, model)
	if _, err = odbi.ormFindInCache(table, model); err != ErrorNotFound {
		return nil, ErrorExist
	}

	row, err := odbi.client.ORM(odbi.db).NewRow(string(table), model)
	if err != nil {
		return nil, err
	}
	insertOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    string(table),
		Row:      row,
		UUIDName: uuid,
	}

	return &OvnCommand{Operations: []libovsdb.Operation{insertOp},
		Exe:     odbi,
		Results: make([][]map[string]interface{}, 1)}, nil
}
func (odbi *ovndb) findTable(mType reflect.Type) TableName {
	for table, tType := range odbi.dbModel.types {
		if tType == mType {
			return table
		}
	}
	return ""
}

func (odbi *ovndb) setUUID(model Model, uuid string) {
	uField := reflect.Indirect(reflect.ValueOf(model)).FieldByName("UUID")
	uField.Set(reflect.ValueOf(uuid))
}

func (odbi *ovndb) getUUID(model Model) string {
	return reflect.Indirect(reflect.ValueOf(model)).FieldByName("UUID").Interface().(string)
}

func (odbi *ovndb) ormEqual(lhs, rhs Model, indexes ...string) (bool, error) {
	api := odbi.client.ORM(odbi.db)
	if lhs.Table() != rhs.Table() {
		return false, nil
	}
	return api.Equal(lhs.Table(), lhs, rhs, indexes...)
}

// ormFindInCache looks for an item in the cache
func (odbi *ovndb) ormFindInCache(table TableName, model Model, index ...string) (Model, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	tableCache, ok := odbi.ormCache[table]
	if !ok {
		return nil, ErrorNotFound
	}

	for _, elem := range tableCache {
		eq, err := odbi.ormEqual(elem.(Model), model, index...)
		if err != nil {
			return nil, err
		}
		if eq {
			return elem.(Model), nil
		}
	}
	return nil, ErrorNotFound
}
