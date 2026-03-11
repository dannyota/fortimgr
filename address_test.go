package fortimgr

import (
	"context"
	"testing"
)

func TestListAddresses(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListAddresses(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/address": `[
				{
					"name": "web-server",
					"type": "ipmask",
					"subnet": ["10.0.0.1", "255.255.255.255"],
					"start-ip": "",
					"end-ip": "",
					"fqdn": "",
					"country": "",
					"wildcard": ["0.0.0.0", "0.0.0.0"],
					"comment": "Production web server",
					"color": 3,
					"associated-interface": "port1"
				},
				{
					"name": "internal-net",
					"type": "ipmask",
					"subnet": "192.168.1.0/255.255.255.0",
					"start-ip": "",
					"end-ip": "",
					"fqdn": "",
					"country": "",
					"wildcard": "",
					"comment": "",
					"color": 0,
					"associated-interface": ""
				},
				{
					"name": "ip-range",
					"type": "iprange",
					"subnet": "",
					"start-ip": "10.0.0.10",
					"end-ip": "10.0.0.20",
					"fqdn": "",
					"country": "",
					"wildcard": "",
					"comment": "DHCP range",
					"color": 0,
					"associated-interface": ""
				},
				{
					"name": "example.com",
					"type": "fqdn",
					"subnet": "",
					"start-ip": "",
					"end-ip": "",
					"fqdn": "example.com",
					"country": "",
					"wildcard": "",
					"comment": "",
					"color": 0,
					"associated-interface": ""
				}
			]`,
		})

		addrs, err := client.ListAddresses(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(addrs) != 4 {
			t.Fatalf("len = %d, want 4", len(addrs))
		}

		// Array subnet → CIDR, /32 stripped.
		a := addrs[0]
		if a.Name != "web-server" {
			t.Errorf("Name = %q", a.Name)
		}
		if a.Type != "ipmask" {
			t.Errorf("Type = %q", a.Type)
		}
		if a.Subnet != "10.0.0.1" {
			t.Errorf("Subnet = %q, want 10.0.0.1 (host, no /32)", a.Subnet)
		}
		if a.Wildcard != "0.0.0.0/0" {
			t.Errorf("Wildcard = %q", a.Wildcard)
		}
		if a.Comment != "Production web server" {
			t.Errorf("Comment = %q", a.Comment)
		}
		if a.Color != 3 {
			t.Errorf("Color = %d", a.Color)
		}
		if a.AssocIntf != "port1" {
			t.Errorf("AssocIntf = %q", a.AssocIntf)
		}

		// String subnet → CIDR.
		if addrs[1].Subnet != "192.168.1.0/24" {
			t.Errorf("Subnet = %q, want 192.168.1.0/24", addrs[1].Subnet)
		}

		// IP range — Type and range fields.
		if addrs[2].Type != "iprange" {
			t.Errorf("Type = %q", addrs[2].Type)
		}
		if addrs[2].StartIP != "10.0.0.10" || addrs[2].EndIP != "10.0.0.20" {
			t.Errorf("StartIP=%q EndIP=%q", addrs[2].StartIP, addrs[2].EndIP)
		}

		// FQDN — Type and value.
		if addrs[3].Type != "fqdn" {
			t.Errorf("Type = %q", addrs[3].Type)
		}
		if addrs[3].FQDN != "example.com" {
			t.Errorf("FQDN = %q", addrs[3].FQDN)
		}
		if addrs[3].Country != "" {
			t.Errorf("Country = %q, want empty", addrs[3].Country)
		}
	})
}

func TestListAddressGroups(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListAddressGroups(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/addrgrp": `[
				{
					"name": "web-servers",
					"member": ["web-01", "web-02", "web-03"],
					"comment": "Web server group",
					"color": 5
				},
				{
					"name": "empty-group",
					"member": [],
					"comment": "",
					"color": 0
				}
			]`,
		})

		groups, err := client.ListAddressGroups(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 2 {
			t.Fatalf("len = %d, want 2", len(groups))
		}

		g := groups[0]
		if g.Name != "web-servers" {
			t.Errorf("Name = %q", g.Name)
		}
		if len(g.Members) != 3 {
			t.Errorf("Members = %v", g.Members)
		}
		if g.Comment != "Web server group" {
			t.Errorf("Comment = %q", g.Comment)
		}
		if g.Color != 5 {
			t.Errorf("Color = %d", g.Color)
		}
	})
}
