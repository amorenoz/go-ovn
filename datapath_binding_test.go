package goovn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var dataPathBindings = map[int](map[string]string){
	1:  {"key1": "val1", "key2": "val2"},
	2:  {"key": "val"},
	20: {"logical-switch": "d9b28e90-379b-4757-aa67-d95f4f7dda6c", "name": "join_ovn-worker2"},
	42: {"the": "answer", "to": "life", ",the": "universe", "and": "everything"},
}

func TestDataPathBinding(t *testing.T) {
	var cmds []*OvnCommand
	ovndbapi := getOVNClient(DBSB)

	t.Logf("Adding DataPathBindings to OVN SB DB")
	for key, extIds := range dataPathBindings {
		cmd, err := ovndbapi.DataPathBindingAdd(key, extIds)
		if err != nil {
			t.Fatal(err)
		}
		cmds = append(cmds, cmd)
	}

	err := ovndbapi.Execute(cmds...)
	if err != nil {
		t.Fatal(err)
	}

	lsdp, err := ovndbapi.DataPathBindingList()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("datapaths: %v", *lsdp[0])

	assert.Equal(t, len(lsdp), len(dataPathBindings), "Number of added DataPathBindings match listed")

	for _, dpb := range lsdp {
		gdpb, err := ovndbapi.DataPathBindingGet(dpb.UUID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, gdpb.ExternalID, dpb.ExternalID, "Get DataPathBinding retuns correct element")
		assert.Equal(t, gdpb.TunnelKey, dpb.TunnelKey, "Get DataPathBinding retuns correct element")
	}

	cmds = make([]*OvnCommand, 0, 2)
	for _, dpb := range lsdp {
		cmd, err := ovndbapi.DataPathBindingDel(dpb.UUID)
		if err != nil {
			t.Fatal(err)
		}
		cmds = append(cmds, cmd)
	}

	err = ovndbapi.Execute(cmds...)
	if err != nil {
		t.Fatal(err)
	}

	lsdp, err = ovndbapi.DataPathBindingList()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, len(lsdp), "After deleting, the number of DataPathBindings is back to 0")
}
