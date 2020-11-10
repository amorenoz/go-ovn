/**
 * Copyright (c) 2020 Red Hat.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

/* PoC state:
TODO:
- Deletion
- Support additional indexes.
	Get(obj {}interface, indexName string, indexValue {})
	or
	Get(obj {}interface, ...{}interface) any number of indexName-indexValue
  It should have a tag?
- Add generic Unit testing
- Specify the supported fields and document them

Open items:
- Should we expose an Add() command
*/

package goovn

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ebay/libovsdb"
)

type OVSDBObj interface {
	GetTable() string
}

// TODO: private?
type OVSObjField struct {
	Name   string
	Column string
	IsRef  bool
	IsUUID bool
	Type   reflect.Type
}

/*GetObjFields returns a list of OVSObjFields for a given Type
It is does so based on the "ovn" tag of the structure
"ovn:{column},[uuid|ref]"
*/
func GetObjFields(objType reflect.Type) []OVSObjField {
	fields := make([]OVSObjField, 0, 0)
	for i := 0; i < objType.NumField(); i++ {
		ref := false
		uuid := false

		field := objType.Field(i)
		tag := field.Tag.Get("ovn")
		if tag == "" {
			continue
		}
		// TODO: Split tag and determine if ref
		tagParts := strings.Split(tag, ",")
		if len(tagParts) > 2 || len(tagParts) == 0 {
			// TODO Handle error
			continue
		} else if len(tagParts) == 2 {
			switch tagParts[1] {
			case "ref":
				ref = true
			case "uuid":
				uuid = true
			}
		}
		col := tagParts[0]

		objField := OVSObjField{
			Name:   field.Name,
			Column: col,
			Type:   field.Type,
			IsRef:  ref,
			IsUUID: uuid,
		}
		fields = append(fields, objField)
	}
	return fields
}

/*
row2Obj: decode a row as a Object based on its tags
*/
func (odbi *ovndb) row2Obj(table string, uuid string, obj interface{}) error {

	objVal := reflect.Indirect(reflect.ValueOf(obj))
	objType := objVal.Type()

	//values := make(map[string]interface{}, 0)

	cacheObj, ok := odbi.cache[table][uuid]
	if !ok {
		return fmt.Errorf("Object with uuid%s not found in cached table %s", uuid, table)
	}

	for _, field := range GetObjFields(objType) {
		fmt.Printf("Evaluating field %v\n", field)
		if field.IsUUID {
			uuidVal := reflect.ValueOf(uuid)
			fmt.Printf("Value %v type %v. Can we set?? %v ", uuidVal, uuidVal.Type().String(), objVal.CanSet())
			// TODO Check the type is string
			objVal.FieldByName(field.Name).Set(uuidVal)
			continue
		}

		if field.Column == "" {
			continue
		}

		rowVal, ok := cacheObj.Fields[field.Column]
		if !ok {
			return fmt.Errorf("Field with column %s not present in Row %v", field.Column, cacheObj.Fields)
		}

		switch field.Type.Kind() {
		case reflect.Int:
			switch rowVal.(type) {
			// FIXME: Verify int64 is correct
			case int:
				intVal := reflect.ValueOf(rowVal)
				fmt.Printf("Value %v type %v", intVal, intVal.Type().String())
				objVal.FieldByName(field.Name).SetInt(intVal.Int())
			default:
				return fmt.Errorf("Type error in field %s, expected int, got %v", field.Column, rowVal)
			}

		case reflect.String:
			switch rowVal.(type) {
			case string:
				objVal.FieldByName(field.Name).SetString(rowVal.(string))
			default:
				return fmt.Errorf("Type error in field %s, expected string, got %v", field.Column, rowVal)
			}
		case reflect.Map:
			fieldValue := objVal.FieldByName(field.Name)
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeMap(field.Type))
			}
			switch rowVal.(type) {
			case libovsdb.OvsMap:
				valMap := rowVal.(libovsdb.OvsMap).GoMap
				for k, v := range valMap {
					kVal := reflect.ValueOf(k)
					vVal := reflect.ValueOf(v)
					if !kVal.Type().ConvertibleTo(field.Type.Key()) {
						return fmt.Errorf("Type error in field %s, map key is of type  %s not convertible to %s", field.Column, kVal.Type().String(), field.Type.Key().String())
					}

					if !vVal.Type().ConvertibleTo(field.Type.Elem()) {
						return fmt.Errorf("Type error in field %s, map key is of type  %s not convertible to %s", field.Column, vVal.Type().String(), field.Type.Elem().String())
					}
					fieldValue.SetMapIndex(kVal.Convert(field.Type.Key()), vVal.Convert(field.Type.Elem()))
				}
			default:
				return fmt.Errorf("Type error in field %s, expected map, got %v", field.Column, rowVal)
			}

		case reflect.Slice:
			if field.Type.Elem().Kind() != reflect.String {
				return fmt.Errorf("Type error in field %s: Only slices of strings are supported", field.Column)
			}
			fieldValue := objVal.FieldByName(field.Name)
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeSlice(field.Type, 0, 0))
			}
			switch rowVal.(type) {
			case string:
				reflect.Append(fieldValue, reflect.ValueOf(rowVal.(string)))
			case []string:
				for str := range rowVal.([]string) {
					reflect.Append(fieldValue, reflect.ValueOf(str))
				}
			case libovsdb.OvsSet:
				for str := range odbi.ConvertGoSetToStringArray(rowVal.(libovsdb.OvsSet)) {
					reflect.Append(fieldValue, reflect.ValueOf(str))
				}
			case libovsdb.UUID:
				reflect.Append(reflect.ValueOf(rowVal.(libovsdb.UUID).GoUUID))
			default:
				return fmt.Errorf("Type error in field %s, expected Set, got %v (%s) ", field.Column, rowVal, reflect.TypeOf(rowVal).String())
			}
		default:
			return fmt.Errorf("Unknown row value type Type error in field %s, expected Ref, got %v (%s) ", field.Column, rowVal, reflect.TypeOf(rowVal).String())
		}
	}

	fmt.Printf("Resulting obj %v", obj)
	return nil
}

func obj2Row(obj interface{}) (OVNRow, error) {
	objVal := reflect.ValueOf(obj)
	objType := objVal.Type()
	objRow := make(OVNRow)

	fmt.Printf("Fields are %v", GetObjFields(objType))

	for _, field := range GetObjFields(objType) {
		if field.Column == "" {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Int:
			objRow[field.Column] = objVal.FieldByName(field.Name).Interface().(int)

		case reflect.String:
			// TODO: Should we support single string refs?
			objRow[field.Column] = objVal.FieldByName(field.Name).String()

		case reflect.Map:
			vMap := objVal.FieldByName(field.Name).Interface().(map[string]string)
			fmt.Printf("vMap is %+v\n, type %s", vMap, reflect.ValueOf(vMap).Type().String())
			oMap, err := libovsdb.NewOvsMap(vMap)
			if err != nil {
				return nil, err
			}
			objRow[field.Column] = oMap

		case reflect.Slice:
			value := objVal.FieldByName(field.Name)

			if field.IsRef {
				valueUUIDs := make([]libovsdb.UUID, 0)
				for i := 0; i < value.Len(); i++ {
					valueUUIDs = append(valueUUIDs, stringToGoUUID(value.Slice(i, i).String()))
				}
				oSet, err := libovsdb.NewOvsSet(valueUUIDs)
				if err != nil {
					return nil, err
				}
				objRow[field.Column] = oSet
			} else {
				oSet, err := libovsdb.NewOvsSet(value.Interface().([]string))
				if err != nil {
					return nil, err
				}
				objRow[field.Column] = oSet
			}

			/// TODO Handle Reference
		}
	}
	return objRow, nil
}

func (odbi *ovndb) List(objPtr interface{}) error {
	objPtrVal := reflect.ValueOf(objPtr)
	if objPtrVal.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("obj parameter must be a pointer to a struct, not %v", objPtrVal.Type().String())
	}
	objSlice := objPtrVal.Elem()
	if objSlice.Type().Kind() != reflect.Slice {
		return fmt.Errorf("obj must be a pointer to a slice of structs")
	}

	// Create a dummy object of the type of elements in the slice to call
	// the ovsdbobj interface. TODO: Figure out a better way to do this
	tempObj := reflect.New(objSlice.Type().Elem())
	tempObjOVS, ok := tempObj.Interface().(OVSDBObj)
	if !ok {
		return fmt.Errorf("obj must implement OVSDBObj interface")
	}
	table := tempObjOVS.GetTable()

	/* TODO: Check against supported tables?*/
	if table == "" {
		return fmt.Errorf("Empty table")
	}

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheTable, ok := odbi.cache[table]
	if !ok {
		return ErrorSchema
	}

	if objSlice.IsNil() {
		objSlice.Set(reflect.MakeSlice(objSlice.Type(), 0, 0))
	}

	for uuid := range cacheTable {
		fmt.Printf("cachetable has %s -> %v\n", uuid, cacheTable[uuid])
		obj := reflect.New(objSlice.Type().Elem())
		err := odbi.row2Obj(table, uuid, obj.Interface())
		fmt.Printf("after row2obj %v", obj)
		if err != nil {
			return err
		}
		objSlice.Set(reflect.Append(objSlice, obj.Elem()))
	}
	return nil
}

/* Get any object based on its reported table name and ovn tags
 */
func (odbi *ovndb) Get(objPtr interface{}, uuid string) error {
	objPtrVal := reflect.ValueOf(objPtr)
	if objPtrVal.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("obj parameter must be a pointer to a struct, not %v", objPtrVal.Type().String())
	}

	obj := objPtrVal.Elem()
	objOVS, ok := obj.Interface().(OVSDBObj)
	if !ok {
		return fmt.Errorf("obj must implement OVSDBObj interface")
	}
	table := objOVS.GetTable()

	return odbi.row2Obj(table, uuid, objPtrVal.Interface())
}

/*
We may want to expose the ovncommands in an AddCmd to externalize transactions
If we do so: the user will do, e.g::
lr := LogicalRouter {...}
 cmd = cli.Add(lr)
 if err := cmd.execute(); err != nil {
    ...
 }

 Another possiblity is to hide that, the potential benefit is that, the ovndbimp
 executes the transaction, we can populate the uuids, so:

lr := LogicalRouter {...}
if err := cli.Add(lr); err != nil {
    fmt.Printf("Added LR with uuid %s\n", lr.UUID)
}
*/
func (odbi *ovndb) Add(objPtr interface{}) (*OvnCommand, error) {
	var operations []libovsdb.Operation

	objPtrVal := reflect.ValueOf(objPtr)
	if objPtrVal.Type().Kind() != reflect.Ptr {
		return nil, fmt.Errorf("obj parameter must be a pointer to a struct, not %v", objPtrVal.Type().String())
	}

	// TODO: Handle already exists

	namedUUID, err := newRowUUID()
	if err != nil {
		return nil, err
	}

	obj := objPtrVal.Elem()
	objRow, err := obj2Row(obj.Interface())
	if err != nil {
		return nil, err
	}

	objOVS, ok := obj.Interface().(OVSDBObj)
	if !ok {
		return nil, fmt.Errorf("obj must implement OVSDBObj interface")
	}
	table := objOVS.GetTable()

	fmt.Printf("objRow %v \n", objRow)
	insertOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    table,
		Row:      objRow,
		UUIDName: namedUUID,
	}
	operations = append(operations, insertOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil

}
