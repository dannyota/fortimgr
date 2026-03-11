package fortimgr

import (
	"context"
	"testing"
)

func TestListVirtualIPs(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListVirtualIPs(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/vip": `[
				{
					"name": "web-vip",
					"extip": "203.0.113.10",
					"mappedip": ["10.0.0.10-10.0.0.10"],
					"extintf": "port1",
					"portforward": "enable",
					"protocol": "tcp",
					"extport": "443",
					"mappedport": "8443",
					"comment": "Web server VIP",
					"color": 4
				},
				{
					"name": "multi-range-vip",
					"extip": "203.0.113.20",
					"mappedip": ["10.0.0.20-10.0.0.25", "10.0.1.20-10.0.1.25"],
					"extintf": "any",
					"portforward": 0,
					"protocol": 0,
					"extport": "0",
					"mappedport": "0",
					"comment": "",
					"color": 0
				}
			]`,
		})

		virtualIPs, err := client.ListVirtualIPs(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(virtualIPs) != 2 {
			t.Fatalf("len = %d, want 2", len(virtualIPs))
		}

		v := virtualIPs[0]
		if v.Name != "web-vip" {
			t.Errorf("Name = %q", v.Name)
		}
		if v.ExtIP != "203.0.113.10" {
			t.Errorf("ExtIP = %q", v.ExtIP)
		}
		if v.MappedIP != "10.0.0.10-10.0.0.10" {
			t.Errorf("MappedIP = %q", v.MappedIP)
		}
		if v.ExtIntf != "port1" {
			t.Errorf("ExtIntf = %q", v.ExtIntf)
		}
		if v.PortForward != "enable" {
			t.Errorf("PortForward = %q", v.PortForward)
		}
		if v.Protocol != "tcp" {
			t.Errorf("Protocol = %q", v.Protocol)
		}
		if v.ExtPort != "443" {
			t.Errorf("ExtPort = %q", v.ExtPort)
		}
		if v.MappedPort != "8443" {
			t.Errorf("MappedPort = %q", v.MappedPort)
		}
		if v.Comment != "Web server VIP" {
			t.Errorf("Comment = %q", v.Comment)
		}
		if v.Color != 4 {
			t.Errorf("Color = %d", v.Color)
		}

		// Multi-range mapped IP → comma-joined.
		v2 := virtualIPs[1]
		if v2.MappedIP != "10.0.0.20-10.0.0.25,10.0.1.20-10.0.1.25" {
			t.Errorf("MappedIP = %q", v2.MappedIP)
		}
		// Enum mapping (int → named string).
		if v2.PortForward != "disable" {
			t.Errorf("PortForward = %q, want \"disable\"", v2.PortForward)
		}
	})
}
