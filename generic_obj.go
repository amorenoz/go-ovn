package goovn

import (
	"fmt"

	"reflect"

	"github.com/ebay/libovsdb"
)

type OVSFieldType string

const (
	OVSTypeInt    OVSFieldType = "Integer"
	OVSTypeString OVSFieldType = "String"
	OVSTypeMap    OVSFieldType = "Map"
	OVSTypeSet    OVSFieldType = "Set"
	OVSTypeRef    OVSFieldType = "Ref" // Accepts a list of uuid strings
)

type OVSDBKey struct {
	Type  OVSFieldType
	Value interface{}
}

type OVSDBTable struct {
	TableName string
	Fields    map[string]OVSFieldType
	/* Indexes are a list of additional indexes (apart from default _uuid) */
	Indexes []string
}

type OVSDBObj struct {
	Table  *OVSDBTable
	Values map[string]interface{}
	UUID   string
}

func (odbi *ovndb) row2Obj(table *OVSDBTable, uuid string) (*OVSDBObj, error) {
	values := make(map[string]interface{}, 0)

	cacheObj, ok := odbi.cache[table.TableName][uuid]
	if !ok {
		return nil, fmt.Errorf("Object with uuid%s not found in cached table %s", uuid, table.TableName)
	}

	for fieldName, fieldType := range table.Fields {
		rowVal, ok := cacheObj.Fields[fieldName]
		if !ok {
			return nil, fmt.Errorf("Field %s not present in Row %v", fieldName, cacheObj.Fields)
		}

		switch fieldType {
		case OVSTypeInt:
			switch rowVal.(type) {
			case int:
				values[fieldName] = rowVal.(int)
			default:
				return nil, fmt.Errorf("Type error in field %s, expected int, got %v", fieldName, rowVal)
			}

		case OVSTypeString:
			switch rowVal.(type) {
			case string:
				values[fieldName] = rowVal.(string)
			default:
				return nil, fmt.Errorf("Type error in field %s, expected string, got %v", fieldName, rowVal)
			}

		case OVSTypeMap:
			values[fieldName] = rowVal.(libovsdb.OvsMap).GoMap

		case OVSTypeSet:
			switch rowVal.(type) {
			case string:
				values[fieldName] = []string{rowVal.(string)}
			case []string:
				values[fieldName] = rowVal
			case libovsdb.OvsSet:
				values[fieldName] = odbi.ConvertGoSetToStringArray(rowVal.(libovsdb.OvsSet))
			default:
				return nil, fmt.Errorf("Type error in field %s, expected Set, got %v (%s) ", fieldName, rowVal, reflect.TypeOf(rowVal).String())
			}
		case OVSTypeRef:
			switch rowVal.(type) {
			case libovsdb.UUID:
				values[fieldName] = []string{rowVal.(libovsdb.UUID).GoUUID}
			case libovsdb.OvsSet:
				values[fieldName] = odbi.ConvertGoSetToStringArray(rowVal.(libovsdb.OvsSet))
			default:
				return nil, fmt.Errorf("Type error in field %s, expected Ref, got %v (%s) ", fieldName, rowVal, reflect.TypeOf(rowVal).String())
			}
		}
	}

	return &OVSDBObj{
		Table:  table,
		UUID:   uuid,
		Values: values,
	}, nil
}

func (odbi *ovndb) ListTable(table *OVSDBTable) ([]*OVSDBObj, error) {
	if table == nil {
		return nil, fmt.Errorf("nil table")
	}
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheTable, ok := odbi.cache[table.TableName]
	if !ok {
		return nil, ErrorSchema
	}
	objList := make([]*OVSDBObj, 0, len(cacheTable))

	for uuid := range cacheTable {
		obj, err := odbi.row2Obj(table, uuid)
		if err != nil {
			return nil, err
		}
		objList = append(objList, obj)
	}
	return objList, nil
}

func (odbi *ovndb) GetObjByUUID(table *OVSDBTable, uuid string) (*OVSDBObj, error) {
	obj, _ := odbi.row2Obj(table, uuid)
	if obj == nil {
		return nil, ErrorNotFound
	}
	return obj, nil
}

func (odbi *ovndb) GetObjByIndex(table *OVSDBTable, index string, value interface{}) ([]*OVSDBObj, error) {
	if table == nil {
		return nil, fmt.Errorf("nil table")
	}

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheTable, ok := odbi.cache[table.TableName]
	if !ok {
		return nil, ErrorSchema
	}
	objList := make([]*OVSDBObj, 0, len(cacheTable))

	for uuid, row := range cacheTable {
		// Only string keys supported
		if rowData, ok := row.Fields[index].(string); ok && rowData == value {
			obj, err := odbi.row2Obj(table, uuid)
			if err != nil {
				return nil, err
			}
			objList = append(objList, obj)
		}
	}
	return objList, nil
}

func (odbi *ovndb) AddObj(obj *OVSDBObj) (*OvnCommand, error) {
	var operations []libovsdb.Operation

	for _, index := range obj.Table.Indexes {
		indexVal, ok := obj.Values[index]
		if !ok {
			return nil, fmt.Errorf("Index value must be added %s", index)
		}

		if obj, _ := odbi.GetObjByIndex(obj.Table, index, indexVal); obj != nil {
			return nil, ErrorExist
		}
	}

	namedUUID, err := newRowUUID()
	if err != nil {
		return nil, err
	}
	objRow := make(OVNRow)

	for fieldName, fieldType := range obj.Table.Fields {
		switch fieldType {
		case OVSTypeInt, OVSTypeString:
			objRow[fieldName] = obj.Values[fieldName]

		case OVSTypeMap:
			if value := obj.Values[fieldName]; value != nil {
				oMap, err := libovsdb.NewOvsMap(value)
				if err != nil {
					return nil, err
				}
				objRow[fieldName] = oMap
			}

		case OVSTypeSet:
			if value := obj.Values[fieldName]; value != nil {
				oSet, err := libovsdb.NewOvsSet(value)
				if err != nil {
					return nil, err
				}
				objRow[fieldName] = oSet
			}
		case OVSTypeRef:
			if value := obj.Values[fieldName]; value != nil {
				valueUUIDs := make([]libovsdb.UUID, 0)
				for _, val := range value.([]string) {
					valueUUIDs = append(valueUUIDs, stringToGoUUID(val))
				}
				oSet, err := libovsdb.NewOvsSet(valueUUIDs)
				if err != nil {
					return nil, err
				}
				objRow[fieldName] = oSet
			}

		}
	}
	insertOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    obj.Table.TableName,
		Row:      objRow,
		UUIDName: namedUUID,
	}
	operations = append(operations, insertOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) DelObj(obj *OVSDBObj) (*OvnCommand, error) {
	var operations []libovsdb.Operation

	// First try deleting by UUID, if not filled, try the first Index
	var condition []interface{}
	if len(obj.UUID) > 0 {
		condition = libovsdb.NewCondition("_uuid", "==", stringToGoUUID(obj.UUID))
	} else {
		if len(obj.Table.Indexes) > 0 {
			condition = libovsdb.NewCondition(obj.Table.Indexes[0], "==",
				obj.Values[obj.Table.Indexes[0]])
		}
	}

	deleteOp := libovsdb.Operation{
		Op:    opDelete,
		Table: obj.Table.TableName,
		Where: []interface{}{condition},
	}
	operations = append(operations, deleteOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}
