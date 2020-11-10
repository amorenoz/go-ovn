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

// LogicalFlow ovnsb item
type LogicalFlow struct {
	UUID            string
	Actions         string
	Match           string
	Pipeline        string
	LogicalDataPath []string
	ExternalID      map[interface{}]interface{}
	Priority        int
	Table           int
}

// LogicalFlowTable
var (
	LFTable = OVSDBTable{
		TableName: TableLogicalFlow,
		Fields: map[string]OVSFieldType{
			"actions":          OVSTypeString,
			"match":            OVSTypeString,
			"pipeline":         OVSTypeString,
			"logical_datapath": OVSTypeRef,
			"external_ids":     OVSTypeMap,
			"priority":         OVSTypeInt,
			"table_id":         OVSTypeInt,
		},
		Indexes: nil,
	}
)

func NewLogicalFlow(obj *OVSDBObj) *LogicalFlow {
	return &LogicalFlow{
		UUID:            obj.UUID,
		Actions:         obj.Values["actions"].(string),
		Match:           obj.Values["match"].(string),
		Pipeline:        obj.Values["pipeline"].(string),
		LogicalDataPath: obj.Values["logical_datapath"].([]string),
		Priority:        obj.Values["priority"].(int),
		Table:           obj.Values["table_id"].(int),
		ExternalID:      obj.Values["external_ids"].(map[interface{}]interface{})}
}

func (odbi *ovndb) lfAddImp(actions string, match string, pipeline string, logical_datapath []string, priority int, table_id int, external_ids map[string]string) (*OvnCommand, error) {
	obj := OVSDBObj{
		Table: &LFTable,
		Values: map[string]interface{}{
			"actions":          actions,
			"match":            match,
			"pipeline":         pipeline,
			"logical_datapath": logical_datapath,
			"priority":         priority,
			"table_id":         table_id,
			"external_ids":     external_ids,
		},
	}
	return odbi.AddObj(&obj)
}

// Delete LogicalFlow by UUID
func (odbi *ovndb) lfDelImp(lf string) (*OvnCommand, error) {
	obj := OVSDBObj{
		Table: &LFTable,
		UUID:  lf,
	}
	return odbi.DelObj(&obj)
}

// Get LogicalFlow by UUID
func (odbi *ovndb) lfGetImp(lf string) (*LogicalFlow, error) {
	obj, err := odbi.GetObjByUUID(&LFTable, lf)
	if err != nil {
		return nil, err
	}
	return NewLogicalFlow(obj), nil
}

func (odbi *ovndb) lfListImp() ([]*LogicalFlow, error) {
	var lfList []*LogicalFlow
	objList, err := odbi.ListTable(&LFTable)
	if err != nil {
		return nil, err
	}
	for _, obj := range objList {
		lfList = append(lfList, NewLogicalFlow(obj))
	}
	return lfList, nil
}
