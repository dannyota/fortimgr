package fortimgr

import (
	"context"
	"testing"
)

func TestListADOMs(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListADOMs(context.Background())
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom": `[
				{
					"name": "root",
					"desc": "Default ADOM",
					"state": 1,
					"mode": 1,
					"os_ver": 7,
					"mr": 4
				},
				{
					"name": "customer-a",
					"desc": "Customer A tenant",
					"state": "enabled",
					"mode": 2,
					"os_ver": 7,
					"mr": 2
				},
				{
					"name": "disabled-adom",
					"desc": "",
					"state": 0,
					"mode": 1,
					"os_ver": 0,
					"mr": 0
				}
			]`,
		})

		adoms, err := client.ListADOMs(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(adoms) != 3 {
			t.Fatalf("len = %d, want 3", len(adoms))
		}

		a := adoms[0]
		if a.Name != "root" {
			t.Errorf("Name = %q", a.Name)
		}
		if a.Desc != "Default ADOM" {
			t.Errorf("Desc = %q", a.Desc)
		}
		if a.State != "enabled" {
			t.Errorf("State = %q, want \"enabled\"", a.State)
		}
		if a.Mode != "normal" {
			t.Errorf("Mode = %q, want \"normal\"", a.Mode)
		}
		if a.OSVer != "7.4" {
			t.Errorf("OSVer = %q, want \"7.4\"", a.OSVer)
		}

		// String enum passthrough.
		if adoms[1].State != "enabled" {
			t.Errorf("State = %q, want \"enabled\"", adoms[1].State)
		}
		if adoms[1].Mode != "backup" {
			t.Errorf("Mode = %q, want \"backup\"", adoms[1].Mode)
		}
		if adoms[1].OSVer != "7.2" {
			t.Errorf("OSVer = %q, want \"7.2\"", adoms[1].OSVer)
		}

		// Disabled ADOM with zero os_ver.
		if adoms[2].State != "disabled" {
			t.Errorf("State = %q, want \"disabled\"", adoms[2].State)
		}
		if adoms[2].OSVer != "" {
			t.Errorf("OSVer = %q, want empty", adoms[2].OSVer)
		}
	})
}
