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

// DataPathBinding ovnsb item
type DataPathBinding struct {
	UUID       string
	TunnelKey  int
	ExternalID map[interface{}]interface{}
}

// DataPathBindingTable
var (
	DPBTable = OVSDBTable{
		TableName: TableDataPathBinding,
		Fields: map[string]OVSFieldType{
			"tunnel_key":   OVSTypeInt,
			"external_ids": OVSTypeMap,
		},
		Indexes: nil,
	}
)

func (odbi *ovndb) dpbAddImp(tunnel_key int, external_ids map[string]string) (*OvnCommand, error) {
	obj := OVSDBObj{
		Table: &DPBTable,
		Values: map[string]interface{}{
			"tunnel_key":   tunnel_key,
			"external_ids": external_ids,
		},
	}
	return odbi.AddObj(&obj)
}

// Delete DataPathBinding by UUID
func (odbi *ovndb) dpbDelImp(dpb string) (*OvnCommand, error) {
	obj := OVSDBObj{
		Table: &DPBTable,
		UUID:  dpb,
	}
	return odbi.DelObj(&obj)
}

// Get DataPathBinding by UUID
func (odbi *ovndb) dpbGetImp(dpb string) (*DataPathBinding, error) {
	obj, err := odbi.GetObjByUUID(&DPBTable, dpb)
	if err != nil {
		return nil, err
	}
	return &DataPathBinding{
		UUID:       obj.UUID,
		TunnelKey:  obj.Values["tunnel_key"].(int),
		ExternalID: obj.Values["external_ids"].(map[interface{}]interface{}),
	}, nil
}

func (odbi *ovndb) dpbListImp() ([]*DataPathBinding, error) {
	var dpbList []*DataPathBinding
	objList, err := odbi.ListTable(&DPBTable)
	if err != nil {
		return nil, err
	}
	for _, obj := range objList {
		dpbList = append(dpbList,
			&DataPathBinding{
				UUID:       obj.UUID,
				TunnelKey:  obj.Values["tunnel_key"].(int),
				ExternalID: obj.Values["external_ids"].(map[interface{}]interface{}),
			})

	}
	return dpbList, nil
}
