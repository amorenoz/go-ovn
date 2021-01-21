package goovn

import (
	"testing"
	//"github.com/stretchr/testify/assert"
)

/*
TODO:
- Test invalid data
- Test empty data
- Test adding multiple elements on a transaction
- Test deletion
*/

func TestDatapathBinding(t *testing.T) {
	// Initialize Client
	client := getOVNClient(DBSB)
	// Get SouthBoundAPI
	api := NewSBApi(client)

	t.Logf("Adding DatapathBindings to OVN SB DB")
	dp := DatapathBinding{
		TunnelKey:   4,
		ExternalIDs: map[string]string{"logical-switch": "d9b28e90-379b-4757-aa67-d95f4f7dda6c", "name": "join_ovn-worker2"},
	}

	cmd, err := api.DatapathBinding.Add(&dp)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Execute(cmd)
	if err != nil {
		t.Fatal(err)
	}

	datas, err := api.DatapathBinding.List()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DatapathBinding list %v", datas)

	uuid := (*datas)[0].UUID
	t.Logf("DatapathBinding UUID %s", uuid)

	data2, err := api.DatapathBinding.Get(uuid)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DatapathBinding retrieved %v", data2)

}
