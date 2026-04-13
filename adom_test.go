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

	// Full fixture: 3 real ADOMs + 2 factory-preset ADOMs the session cannot access.
	forwardFixtures := map[string]string{
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
			},
			{
				"name": "FortiAnalyzer",
				"desc": "",
				"state": 1,
				"mode": 1,
				"os_ver": 7,
				"mr": 0
			},
			{
				"name": "FortiMail",
				"desc": "",
				"state": 1,
				"mode": 1,
				"os_ver": 7,
				"mr": 0
			}
		]`,
	}

	// Session scope: current ADOM is "root", plus "customer-a" and "disabled-adom" accessible.
	// The factory presets (FortiAnalyzer, FortiMail) are not in scope.
	proxyFixtures := map[string]string{
		"/gui/sys/config": `{
			"admin_user_name": "admin",
			"adom": {"name": "root"},
			"adoms": ["customer-a", "disabled-adom"]
		}`,
	}

	t.Run("filtered to session scope", func(t *testing.T) {
		client := newTestClientWithProxy(t, forwardFixtures, proxyFixtures)

		adoms, err := client.ListADOMs(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(adoms) != 3 {
			t.Fatalf("len = %d, want 3 (factory presets should be filtered)", len(adoms))
		}

		got := map[string]ADOM{}
		for _, a := range adoms {
			got[a.Name] = a
		}
		for _, name := range []string{"root", "customer-a", "disabled-adom"} {
			if _, ok := got[name]; !ok {
				t.Errorf("missing %q in filtered result", name)
			}
		}
		for _, name := range []string{"FortiAnalyzer", "FortiMail"} {
			if _, ok := got[name]; ok {
				t.Errorf("factory preset %q leaked through filter", name)
			}
		}

		// Spot-check field decoding still works after filtering.
		a := got["root"]
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

		if got["customer-a"].Mode != "backup" || got["customer-a"].OSVer != "7.2" {
			t.Errorf("customer-a = %+v", got["customer-a"])
		}
		if got["disabled-adom"].State != "disabled" {
			t.Errorf("disabled-adom state = %q", got["disabled-adom"].State)
		}
	})

	t.Run("all=true returns global list", func(t *testing.T) {
		client := newTestClientWithProxy(t, forwardFixtures, proxyFixtures)

		adoms, err := client.ListADOMs(context.Background(), true)
		if err != nil {
			t.Fatal(err)
		}
		if len(adoms) != 5 {
			t.Fatalf("len = %d, want 5 (all ADOMs including factory presets)", len(adoms))
		}
	})

	t.Run("all=false is equivalent to default", func(t *testing.T) {
		client := newTestClientWithProxy(t, forwardFixtures, proxyFixtures)

		adoms, err := client.ListADOMs(context.Background(), false)
		if err != nil {
			t.Fatal(err)
		}
		if len(adoms) != 3 {
			t.Fatalf("len = %d, want 3", len(adoms))
		}
	})
}
