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

package goovn

import (
	"fmt"

	"github.com/ebay/libovsdb"
)

// DataPathBinding ovnsb item
type DataPathBinding struct {
	UUID       string
	TunnelKey  int
	ExternalID map[interface{}]interface{}
}

func (odbi *ovndb) dpbAddImp(tunnel_key int, external_ids map[string]string) (*OvnCommand, error) {
	var operations []libovsdb.Operation

	namedUUID, err := newRowUUID()
	if err != nil {
		return nil, err
	}

	dpbRow := make(OVNRow)
	dpbRow["tunnel_key"] = tunnel_key

	if external_ids != nil {
		oMap, err := libovsdb.NewOvsMap(external_ids)
		if err != nil {
			return nil, err
		}
		dpbRow["external_ids"] = oMap
	}

	insertDataPathBindingOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    TableDataPathBinding,
		Row:      dpbRow,
		UUIDName: namedUUID,
	}

	operations = append(operations, insertDataPathBindingOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

// Delete DataPathBinding by UUID
func (odbi *ovndb) dpbDelImp(dpb string) (*OvnCommand, error) {
	var operations []libovsdb.Operation

	condition := libovsdb.NewCondition("_uuid", "==", stringToGoUUID(dpb))
	deleteOp := libovsdb.Operation{
		Op:    opDelete,
		Table: TableDataPathBinding,
		Where: []interface{}{condition},
	}
	operations = append(operations, deleteOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

// Get DataPathBinding by tunnel_key
func (odbi *ovndb) dpbGetImp(dpb string) ([]*DataPathBinding, error) {
	var dpbList []*DataPathBinding
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheDataPathBinding, ok := odbi.cache[TableDataPathBinding]
	if !ok {
		return nil, ErrorNotFound
	}

	for uuid, _ := range cacheDataPathBinding {
		if uuid == dpb {
			dpb, err := odbi.rowToDataPathBinding(uuid)
			if err != nil {
				return nil, err
			}
			dpbList = append(dpbList, dpb)
		}
	}

	if len(dpbList) == 0 {
		return nil, ErrorNotFound
	}
	return dpbList, nil
}

func (odbi *ovndb) rowToDataPathBinding(uuid string) (*DataPathBinding, error) {
	cacheDPB, ok := odbi.cache[TableDataPathBinding][uuid]
	if !ok {
		return nil, fmt.Errorf("DataPathBinding with uuid%s not found", uuid)
	}

	dpb := &DataPathBinding{
		UUID:       uuid,
		TunnelKey:  cacheDPB.Fields["tunnel_key"].(int),
		ExternalID: cacheDPB.Fields["external_ids"].(libovsdb.OvsMap).GoMap,
	}

	return dpb, nil
}

func (odbi *ovndb) dpbListImp() ([]*DataPathBinding, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheDPB, ok := odbi.cache[TableDataPathBinding]
	if !ok {
		return nil, ErrorSchema
	}

	listDPB := make([]*DataPathBinding, 0, len(cacheDPB))

	for uuid := range cacheDPB {
		dpb, err := odbi.rowToDataPathBinding(uuid)
		if err != nil {
			return nil, err
		}
		listDPB = append(listDPB, dpb)
	}
	return listDPB, nil
}
