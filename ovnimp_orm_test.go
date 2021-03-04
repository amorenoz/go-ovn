package goovn

import (
	"testing"
)

// For this test we don't need all these fields
// This type can be autogenerated
type testLS struct {
	UUID             string            `ovs:"_uuid"`
	QosRules         []string          `ovs:"qos_rules"`
	OtherConfig      map[string]string `ovs:"other_config"`
	ExternalIds      map[string]string `ovs:"external_ids"`
	ForwardingGroups []string          `ovs:"forwarding_groups"`
	Name             string            `ovs:"name"`
	Ports            []string          `ovs:"ports"`
	Acls             []string          `ovs:"acls"`
	LoadBalancer     []string          `ovs:"load_balancer"`
	DnsRecords       []string          `ovs:"dns_records"`
}

func (*testLS) Table() TableName { return "Logical_Switch" }

// For this test we don't need all these fields
// This type can be autogenerated
type testLSP struct {
	UUID             string            `ovs:"_uuid"`
	Type             string            `ovs:"type"`
	TagRequest       []int             `ovs:"tag_request"`
	Tag              []int             `ovs:"tag"`
	DynamicAddresses []string          `ovs:"dynamic_addresses"`
	PortSecurity     []string          `ovs:"port_security"`
	Enabled          []bool            `ovs:"enabled"`
	Name             string            `ovs:"name"`
	Up               []bool            `ovs:"up"`
	Dhcpv4Options    []string          `ovs:"dhcpv4_options"`
	HaChassisGroup   []string          `ovs:"ha_chassis_group"`
	ExternalIds      map[string]string `ovs:"external_ids"`
	Dhcpv6Options    []string          `ovs:"dhcpv6_options"`
	Options          map[string]string `ovs:"options"`
	ParentName       []string          `ovs:"parent_name"`
	Addresses        []string          `ovs:"addresses"`
}

func (*testLSP) Table() TableName { return "Logical_Switch_Port" }

// This simple test just adds a Logical Switch and a Logical Port associated with it
// More than a Unit Test, it's a demonstrator
func TestORMUpdate(t *testing.T) {
	DBmodel, err := NewDBModel([]Model{&testLS{}, &testLSP{}})
	if err != nil {
		t.Fatal(err)
	}
	orm := getOVNClientORM(DBNB, DBmodel)

	// Add a Logical Switch
	ls1 := testLS{
		Name:        "someswitch",
		ExternalIds: map[string]string{"foo": "bar"},
	}
	createLSOp, err := orm.Create(&ls1)
	if err != nil {
		t.Error(err)
	}
	err = orm.Execute(createLSOp)
	if err != nil {
		t.Error(err)
	}

	// Reread ls1 to get UUID. TODO: have Execute (or even Create()) return the created UUID?
	if err := orm.Get(&ls1, "name"); err != nil {
		t.Error(err)
	}

	// Add a Logical Switch Port
	uuid, err := newRowUUID()
	if err != nil {
		t.Fatal(err)
	}
	lsp1 := testLSP{
		UUID: uuid,
		Name: "someport",
	}
	createLSPOp, err := orm.Create(&lsp1)
	if err != nil {
		t.Error(err)
	}

	// Update LSP to point to LSP
	ls1.Ports = append(ls1.Ports, lsp1.UUID)
	updateOp, err := orm.Update(&ls1)
	if err != nil {
		t.Error(err)
	}

	err = orm.Execute(createLSPOp, updateOp)
	if err != nil {
		t.Error(err)
	}

	// Verify:
	var switches []testLS
	var sports []testLSP
	err = orm.List(&switches)
	err = orm.List(&sports)
	t.Logf("Logical Switches: %+v", switches)
	t.Logf("Logical Switch Ports: %+v", sports)

	if len(switches) != 1 {
		t.Fatal("Switch not added")
	}
	if len(sports) != 1 {
		t.Fatal("Switch Port not added")
	}
	if switches[0].Ports[0] != sports[0].UUID {
		t.Errorf("Switch not linked to Port")
	}
}
