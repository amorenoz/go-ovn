/**
 * Copyright (c) 2017 eBay Inc.
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
	"github.com/ebay/libovsdb"
)

// LogicalRouter ovnnb item
type LogicalRouter struct {
	UUID    string
	Name    string
	Enabled []bool

	Ports        []string
	StaticRoutes []string
	NAT          []string
	LoadBalancer []string

	Options    map[string]string
	ExternalID map[string]string
}

func (odbi *ovndb) lrAddImp(name string, external_ids map[string]string) (*OvnCommand, error) {
	namedUUID, err := newRowUUID()
	if err != nil {
		return nil, err
	}
	row := make(map[string]interface{})
	row["name"] = name

	if external_ids != nil {
		row["external_ids"] = external_ids
	}

	if uuid := odbi.getRowUUID(TableLogicalRouter, row); len(uuid) > 0 {
		return nil, ErrorExist
	}

	ovsRow, err := odbi.Api().NewRow(TableLogicalRouter, row)
	if err != nil {
		return nil, err
	}

	insertOp := libovsdb.Operation{
		Op:       opInsert,
		Table:    TableLogicalRouter,
		Row:      ovsRow,
		UUIDName: namedUUID,
	}

	operations := []libovsdb.Operation{insertOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) lrDelImp(name string) (*OvnCommand, error) {
	condition, err := odbi.Api().NewCondition(TableLogicalRouter, "name", "==", name)
	if err != nil {
		return nil, err
	}
	deleteOp := libovsdb.Operation{
		Op:    opDelete,
		Table: TableLogicalRouter,
		Where: []interface{}{condition},
	}
	operations := []libovsdb.Operation{deleteOp}
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) lrGetImp(name string) ([]*LogicalRouter, error) {
	var lrList []*LogicalRouter

	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheLogicalRouter, ok := odbi.cache[TableLogicalRouter]
	if !ok {
		return nil, ErrorNotFound
	}

	for uuid, drows := range cacheLogicalRouter {
		if lrName, ok := drows.Fields["name"].(string); ok && lrName == name {
			lr := odbi.rowToLogicalRouter(uuid)
			lrList = append(lrList, lr)
		}
	}
	return lrList, nil
}

func (odbi *ovndb) rowToLogicalRouter(uuid string) *LogicalRouter {
	cacheLogicalRouter, ok := odbi.cache[TableLogicalRouter][uuid]
	if !ok {
		return nil
	}
	var data map[string]interface{}
	if err := odbi.Api().GetRowData(TableLogicalRouter, &cacheLogicalRouter, &data); err != nil {
		return nil
	}

	lr := &LogicalRouter{
		UUID:         uuid,
		Name:         data["name"].(string),
		Options:      data["options"].(map[string]string),
		ExternalID:   data["external_ids"].(map[string]string),
		Enabled:      data["enabled"].([]bool),
		LoadBalancer: data["load_balancer"].([]string),
		Ports:        data["ports"].([]string),
		StaticRoutes: data["static_routes"].([]string),
		NAT:          data["nat"].([]string),
	}
	return lr
}

// Get all logical routers
func (odbi *ovndb) lrListImp() ([]*LogicalRouter, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheLogicalRouter, ok := odbi.cache[TableLogicalRouter]
	if !ok {
		return nil, ErrorNotFound
	}

	listLR := make([]*LogicalRouter, 0, len(cacheLogicalRouter))
	for uuid := range cacheLogicalRouter {
		listLR = append(listLR, odbi.rowToLogicalRouter(uuid))
	}

	return listLR, nil
}

func (odbi *ovndb) lrlbAddImp(lr string, lb string) (*OvnCommand, error) {
	var operations []libovsdb.Operation
	row := make(OVNRow)
	row["name"] = lb
	lbuuid := odbi.getRowUUID(TableLoadBalancer, row)
	if len(lbuuid) == 0 {
		return nil, ErrorNotFound
	}

	mutation, err := odbi.Api().NewMutation(TableLogicalRouter, "load_balancer", opInsert, []string{lbuuid})
	if err != nil {
		return nil, err
	}
	row = make(OVNRow)
	row["name"] = lr
	lruuid := odbi.getRowUUID(TableLogicalRouter, row)
	if len(lruuid) == 0 {
		return nil, ErrorNotFound
	}
	condition, err := odbi.Api().NewCondition(TableLogicalRouter, "name", "==", lr)
	if err != nil {
		return nil, err
	}

	mutateOp := libovsdb.Operation{
		Op:        opMutate,
		Table:     TableLogicalRouter,
		Mutations: []interface{}{mutation},
		Where:     []interface{}{condition},
	}
	operations = append(operations, mutateOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) lrlbDelImp(lr string, lb string) (*OvnCommand, error) {
	var operations []libovsdb.Operation
	row := make(OVNRow)
	row["name"] = lb
	lbuuid := odbi.getRowUUID(TableLoadBalancer, row)
	if len(lbuuid) == 0 {
		return nil, ErrorNotFound
	}

	row = make(OVNRow)
	row["name"] = lr
	lruuid := odbi.getRowUUID(TableLogicalRouter, row)
	if len(lruuid) == 0 {
		return nil, ErrorNotFound
	}

	mutation, err := odbi.Api().NewMutation(TableLogicalRouter, "load_balancer", opDelete, []string{lbuuid})
	if err != nil {
		return nil, err
	}
	// mutate  lswitch for the corresponding load_balancer
	mucondition, err := odbi.Api().NewCondition(TableLogicalRouter, "name", "==", lr)
	if err != nil {
		return nil, err
	}
	mutateOp := libovsdb.Operation{
		Op:        opMutate,
		Table:     TableLogicalRouter,
		Mutations: []interface{}{mutation},
		Where:     []interface{}{mucondition},
	}
	operations = append(operations, mutateOp)
	return &OvnCommand{operations, odbi, make([][]map[string]interface{}, len(operations))}, nil
}

func (odbi *ovndb) lrlbListImp(lr string) ([]*LoadBalancer, error) {
	odbi.cachemutex.RLock()
	defer odbi.cachemutex.RUnlock()

	cacheLogicalRouter, ok := odbi.cache[TableLogicalRouter]
	if !ok {
		return nil, ErrorSchema
	}
	for _, drows := range cacheLogicalRouter {
		if router, ok := drows.Fields["name"].(string); ok && router == lr {
			var lrdata map[string]interface{}
			if err := odbi.Api().GetRowData(TableLogicalRouter, &drows, &lrdata); err != nil {
				return nil, err
			}
			lbs := lrdata["load_balancer"].([]string)
			listLB := make([]*LoadBalancer, 0, len(lbs))
			for _, l := range lbs {
				lb, err := odbi.rowToLB(l)
				if err != nil {
					return nil, err
				}
				listLB = append(listLB, lb)
			}
			return listLB, nil
			return []*LoadBalancer{}, nil
		}
	}
	return nil, ErrorNotFound
}
