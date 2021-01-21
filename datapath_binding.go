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

import ()

/*PoC state:
TODO:
Autogenerate this using database schema
*/

// DataPathBindingTable
var (
	DPBTable string = TableDataPathBinding
)

// DataPathBinding ovnsb item
type DatapathBinding struct {
	UUID        string            `ovn:",uuid"`
	TunnelKey   int               `ovn:"tunnel_key"`
	ExternalIDs map[string]string `ovn:"external_ids"`
}

func (DatapathBinding) GetTable() string {
	return DPBTable
}

type DatapathBindingApi struct {
	client Client
}

func (a DatapathBindingApi) Get(uuid string) (*DatapathBinding, error) {
	var obj DatapathBinding
	if err := a.client.Get(&obj, uuid); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (a DatapathBindingApi) List() (*[]DatapathBinding, error) {
	var obj []DatapathBinding
	if err := a.client.List(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (a DatapathBindingApi) Add(obj *DatapathBinding) (*OvnCommand, error) {
	return a.client.Add(obj)
}

type SBApi struct {
	DatapathBinding DatapathBindingApi
}

func NewSBApi(c Client) SBApi {
	return SBApi{
		DatapathBinding: DatapathBindingApi{
			client: c,
		},
	}
}
