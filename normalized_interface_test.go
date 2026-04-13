package fortimgr

import (
	"context"
	"testing"
)

func TestListNormalizedInterfaces(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListNormalizedInterfaces(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClient(t, map[string]string{})
		_, err := client.ListNormalizedInterfaces(context.Background(), "bad adom")
		if err == nil {
			t.Fatal("want ErrInvalidName, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		// Fixture covers:
		//  - A mapped interface with a single scope (fortilink on fw-01)
		//  - A mapped interface with two scopes (wan1 on fw-01 and fw-02)
		//    → must fan out into 2 Mappings entries
		//  - An unmapped declaration (no dynamic_mapping at all)
		//  - A zone-only entry (wildcard=1)
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/dynamic/interface": `[
				{
					"name": "fortilink",
					"oid": 100,
					"single-intf": 1,
					"zone-only": 0,
					"wildcard": 0,
					"default-mapping": 0,
					"color": 0,
					"dynamic_mapping": [
						{
							"_scope": [{"name": "fw-01", "vdom": "root"}],
							"local-intf": ["fortilink"],
							"intrazone-deny": 0
						}
					]
				},
				{
					"name": "wan1",
					"oid": 101,
					"single-intf": 1,
					"zone-only": 0,
					"wildcard": 0,
					"default-mapping": 0,
					"color": 2,
					"dynamic_mapping": [
						{
							"_scope": [
								{"name": "fw-01", "vdom": "root"},
								{"name": "fw-02", "vdom": "root"}
							],
							"local-intf": ["wan1", "wan1-bak"],
							"intrazone-deny": 1
						}
					]
				},
				{
					"name": "unmapped-decl",
					"oid": 102,
					"single-intf": 1,
					"zone-only": 0,
					"wildcard": 0,
					"default-mapping": 0,
					"color": 0
				},
				{
					"name": "dmz-zone",
					"oid": 103,
					"single-intf": 0,
					"zone-only": 1,
					"wildcard": 1,
					"default-mapping": 0,
					"color": 5
				}
			]`,
		})

		ifaces, err := client.ListNormalizedInterfaces(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(ifaces) != 4 {
			t.Fatalf("len = %d, want 4", len(ifaces))
		}

		// fortilink — one scope, one mapping
		i0 := ifaces[0]
		if i0.Name != "fortilink" || !i0.SingleIntf || i0.ZoneOnly || i0.Wildcard {
			t.Errorf("fortilink flags wrong: %+v", i0)
		}
		if len(i0.Mappings) != 1 {
			t.Fatalf("fortilink Mappings = %d, want 1", len(i0.Mappings))
		}
		m := i0.Mappings[0]
		if m.Device != "fw-01" || m.VDOM != "root" {
			t.Errorf("Mapping scope = %q/%q", m.Device, m.VDOM)
		}
		if len(m.LocalIntf) != 1 || m.LocalIntf[0] != "fortilink" {
			t.Errorf("LocalIntf = %v", m.LocalIntf)
		}
		if m.IntrazoneDeny {
			t.Errorf("IntrazoneDeny should be false")
		}

		// wan1 — two scopes inside one mapping, fan out into 2 Mappings entries
		i1 := ifaces[1]
		if len(i1.Mappings) != 2 {
			t.Fatalf("wan1 Mappings = %d, want 2 (fanned out from 2 scopes)", len(i1.Mappings))
		}
		if i1.Mappings[0].Device != "fw-01" || i1.Mappings[1].Device != "fw-02" {
			t.Errorf("fan-out order wrong: %+v, %+v", i1.Mappings[0], i1.Mappings[1])
		}
		// LocalIntf must be shared across both fan-out entries
		for _, fm := range i1.Mappings {
			if len(fm.LocalIntf) != 2 || fm.LocalIntf[0] != "wan1" || fm.LocalIntf[1] != "wan1-bak" {
				t.Errorf("wan1 LocalIntf = %v", fm.LocalIntf)
			}
			if !fm.IntrazoneDeny {
				t.Errorf("wan1 IntrazoneDeny should be true")
			}
		}

		// Unmapped declaration — Mappings is nil
		i2 := ifaces[2]
		if i2.Name != "unmapped-decl" {
			t.Errorf("Name = %q", i2.Name)
		}
		if i2.Mappings != nil {
			t.Errorf("Mappings should be nil for unmapped decl, got %+v", i2.Mappings)
		}

		// Zone-only entry
		i3 := ifaces[3]
		if !i3.ZoneOnly || !i3.Wildcard {
			t.Errorf("dmz-zone flags: ZoneOnly=%v Wildcard=%v", i3.ZoneOnly, i3.Wildcard)
		}
		if i3.Color != 5 {
			t.Errorf("Color = %d, want 5", i3.Color)
		}
	})
}
