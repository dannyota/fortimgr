package fortimgr

import (
	"context"
	"errors"
	"testing"
)

func TestListDeviceAssignedPackages(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListDeviceAssignedPackages(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid adom", func(t *testing.T) {
		client := newTestClientWithProxy(t, map[string]string{
			"/dvmdb/adom": `[]`,
		}, nil)
		_, err := client.ListDeviceAssignedPackages(context.Background(), "bad/adom")
		if !errors.Is(err, ErrInvalidName) {
			t.Errorf("err = %v, want ErrInvalidName", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, map[string]string{
			"/dvmdb/adom": `[{"name":"root","oid":7}]`,
		}, map[string]string{
			"/gui/adoms/7/devices/assignedpkgs:get": `{
				"456:2": {
					"deviceOid": 456,
					"vdomOid": 2,
					"pkg": {"name": "pkg-b", "oid": 90, "flags": 3, "status": 1},
					"profileDirty": true
				},
				"123:1": {
					"deviceOid": 123,
					"vdomOid": 1,
					"pkg": {"name": "pkg-a", "oid": 89, "flags": 1, "status": 2},
					"fap_prof": {"name": "fap", "oid": 10},
					"fext_prof": {"name": "fext", "oid": 11}
				}
			}`,
		})

		items, err := client.ListDeviceAssignedPackages(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 2 {
			t.Fatalf("len = %d, want 2", len(items))
		}
		if items[0].DeviceOID != 123 || items[0].VDOMOID != 1 || items[0].Package.Name != "pkg-a" {
			t.Errorf("items[0] = %+v", items[0])
		}
		if items[0].FAPProfile.Name != "fap" || items[0].FExtProfile.Name != "fext" {
			t.Errorf("profile refs = %+v %+v", items[0].FAPProfile, items[0].FExtProfile)
		}
		if !items[1].ProfileDirty {
			t.Errorf("items[1].ProfileDirty = false, want true")
		}
	})
}
