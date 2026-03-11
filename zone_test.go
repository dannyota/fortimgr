package fortimgr

import (
	"context"
	"testing"
)

func TestListZones(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListZones(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/system/zone": `[
				{
					"name": "trust",
					"interface": ["port1", "port2", "port3"],
					"intrazone": 0,
					"description": "Trusted internal zone"
				},
				{
					"name": "untrust",
					"interface": ["wan1"],
					"intrazone": "deny",
					"description": ""
				}
			]`,
		})

		zones, err := client.ListZones(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(zones) != 2 {
			t.Fatalf("len = %d, want 2", len(zones))
		}

		z := zones[0]
		if z.Name != "trust" {
			t.Errorf("Name = %q", z.Name)
		}
		if len(z.Interfaces) != 3 || z.Interfaces[0] != "port1" {
			t.Errorf("Interfaces = %v", z.Interfaces)
		}
		if z.Intrazone != "allow" {
			t.Errorf("Intrazone = %q, want \"allow\"", z.Intrazone)
		}
		if z.Description != "Trusted internal zone" {
			t.Errorf("Description = %q", z.Description)
		}

		if zones[1].Name != "untrust" {
			t.Errorf("Name = %q", zones[1].Name)
		}
		if zones[1].Intrazone != "deny" {
			t.Errorf("Intrazone = %q, want \"deny\"", zones[1].Intrazone)
		}
		if len(zones[1].Interfaces) != 1 || zones[1].Interfaces[0] != "wan1" {
			t.Errorf("Interfaces = %v", zones[1].Interfaces)
		}
	})
}
