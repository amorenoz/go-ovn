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
	return nil, nil
}

// Delete DataPathBinding by tunnel_key
func (odbi *ovndb) dpbDelImp(key int) (*OvnCommand, error) {
	return nil, nil
}

// Get DataPathBinding by tunnel_key
func (odbi *ovndb) dpbGetImp(key int) ([]*DataPathBinding, error) {
	return nil, nil
}

func (odbi *ovndb) rowToDataPathBinding(uuid string) (*DataPathBinding, error) {
	return nil, nil
}

func (odbi *ovndb) dpbListImp() ([]*DataPathBinding, error) {
	return nil, fmt.Errorf("not impl")
}
