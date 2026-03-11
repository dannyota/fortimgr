package fortimgr

import (
	"context"
	"testing"
)

func TestListInterfaces(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListInterfaces(context.Background(), "fw-01", "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/device/fw-01/vdom/root/system/interface": `[
				{
					"name": "port1",
					"ip": ["10.0.1.1", "255.255.255.0"],
					"type": 0,
					"status": 1,
					"role": 0,
					"mode": 0,
					"allowaccess": "ping https ssh",
					"vdom": "root",
					"zone": "trust",
					"vlanid": 0,
					"mtu": 1500,
					"speed": "auto",
					"alias": "LAN",
					"description": "Internal LAN interface"
				},
				{
					"name": "vlan100",
					"ip": ["192.168.100.1", "255.255.255.255"],
					"type": 1,
					"status": 0,
					"role": "wan",
					"mode": 1,
					"allowaccess": ["ping", "https"],
					"vdom": "root",
					"zone": "",
					"vlanid": 100,
					"mtu": 9000,
					"speed": "1000full",
					"alias": "",
					"description": ""
				}
			]`,
		})

		interfaces, err := client.ListInterfaces(context.Background(), "fw-01", "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(interfaces) != 2 {
			t.Fatalf("len = %d, want 2", len(interfaces))
		}

		iface := interfaces[0]
		if iface.Name != "port1" {
			t.Errorf("Name = %q", iface.Name)
		}
		if iface.IP != "10.0.1.1/24" {
			t.Errorf("IP = %q, want \"10.0.1.1/24\"", iface.IP)
		}
		if iface.Type != "physical" {
			t.Errorf("Type = %q, want \"physical\"", iface.Type)
		}
		if iface.Status != "up" {
			t.Errorf("Status = %q, want \"up\"", iface.Status)
		}
		if iface.Role != "lan" {
			t.Errorf("Role = %q, want \"lan\"", iface.Role)
		}
		if iface.Mode != "static" {
			t.Errorf("Mode = %q, want \"static\"", iface.Mode)
		}
		if iface.AllowAccess != "ping https ssh" {
			t.Errorf("AllowAccess = %q, want \"ping https ssh\"", iface.AllowAccess)
		}
		if iface.VDOM != "root" {
			t.Errorf("VDOM = %q", iface.VDOM)
		}
		if iface.Zone != "trust" {
			t.Errorf("Zone = %q", iface.Zone)
		}
		if iface.MTU != 1500 {
			t.Errorf("MTU = %d", iface.MTU)
		}
		if iface.Alias != "LAN" {
			t.Errorf("Alias = %q", iface.Alias)
		}

		iface2 := interfaces[1]
		if iface2.IP != "192.168.100.1" {
			t.Errorf("IP = %q, want \"192.168.100.1\" (host, no /32)", iface2.IP)
		}
		if iface2.Type != "vlan" {
			t.Errorf("Type = %q, want \"vlan\"", iface2.Type)
		}
		if iface2.Status != "down" {
			t.Errorf("Status = %q, want \"down\"", iface2.Status)
		}
		if iface2.Role != "wan" {
			t.Errorf("Role = %q, want \"wan\"", iface2.Role)
		}
		if iface2.Mode != "dhcp" {
			t.Errorf("Mode = %q, want \"dhcp\"", iface2.Mode)
		}
		if iface2.AllowAccess != "ping https" {
			t.Errorf("AllowAccess = %q, want \"ping https\" (from array)", iface2.AllowAccess)
		}
		if iface2.VlanID != 100 {
			t.Errorf("VlanID = %d, want 100", iface2.VlanID)
		}
		if iface2.MTU != 9000 {
			t.Errorf("MTU = %d, want 9000", iface2.MTU)
		}
	})
}
