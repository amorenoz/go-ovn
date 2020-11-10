package goovn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalFlow(t *testing.T) {
	var cmd *OvnCommand
	var err error
	var actions, match, pipeline string
	var logicalDPRef []string
	var extID map[string]string
	var prio, table int

	ovndbapi := getOVNClient(DBSB)

	t.Logf("Adding a DataPathBinding")
	cmd, err = ovndbapi.DataPathBindingAdd(2, map[string]string{"logical-switch": "d9b28e90-379b-4757-aa67-d95f4f7dda6c", "name": "join_ovn-worker2"})
	if err != nil {
		t.Fatal(err)
	}
	err = ovndbapi.Execute(cmd)

	datapaths, err := ovndbapi.DataPathBindingList()
	datapathUUID := datapaths[0].UUID

	t.Logf("Adding LogicalFlows")
	t.Logf("Adding a valid LogicalFlow")
	actions = "arp { eth.dst = ff:ff:ff:ff:ff:ff; arp.spa = reg1; arp.tpa = reg0; arp.op = 1; output; };"
	match = "eth.dst == 00:00:00:00:00:00 && ip4"
	pipeline = "ingress"
	prio = 100
	table = 14
	extID = map[string]string{"source": "ovn-northd.c:11273", "stage-name": "lr_in_arp_request"}
	logicalDPRef = append(logicalDPRef, datapathUUID)

	cmd, err = ovndbapi.LogicalFlowAdd(actions, match, pipeline, logicalDPRef, prio, table, extID)
	assert.NoError(t, err, "Adding a logical flow must succeed")
	err = ovndbapi.Execute(cmd)
	assert.NoError(t, err, "Adding a logical flow must succeed")

	t.Logf("Listing LogicalFlows")
	lfs, err := ovndbapi.LogicalFlowList()
	assert.NoError(t, err, "And does not generate error")
	if !assert.Len(t, lfs, 1, "List yields one element after adding it") {
		panic("Cannot proceed with test")
	}

	lf, err := ovndbapi.LogicalFlowGet(lfs[0].UUID)
	assert.Equal(t, lf, lfs[0], "Can get logical flows by uuid")
	assert.NoError(t, err, "And does not generate error")

	assert.Equal(t, lf.Actions, actions, "Logical Flow has correct actions")
	assert.Equal(t, lf.Match, match, "Logical Flow has correct match")
	assert.Equal(t, lf.Pipeline, pipeline, "Logical Flow has correct pipeline")
	assert.Equal(t, lf.LogicalDataPath, logicalDPRef, "Logical Flow has correct logical datapath")
	for k, val := range extID {
		assert.Equal(t, lf.ExternalID[k].(string), val, "Logical Flow has correct extIDS")
	}
	assert.Equal(t, lf.Priority, prio, "Logical Flow has correct prio")
	assert.Equal(t, lf.Table, table, "Logical Flow has correct table")

	cmd, err = ovndbapi.LogicalFlowDel(lf.UUID)
	assert.NoError(t, err, "Deleting a logical flow must succeed")
	err = ovndbapi.Execute(cmd)
	assert.NoError(t, err, "Deleting a logical flow must succeed")

	lfs, err = ovndbapi.LogicalFlowList()
	assert.Len(t, lfs, 0, "List yields no elements after deletion")
	assert.NoError(t, err, "And does not generate error")

}
