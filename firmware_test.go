package fortimgr

import (
	"context"
	"testing"
)

func TestListDeviceFirmware(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListDeviceFirmware(context.Background())
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/adom/dvm/firmware/management": `[
				{
					"name": "FortiOS v7.4",
					"isGroup": 1,
					"devsn": "",
					"platform_str": "",
					"curr_ver": "",
					"curr_build": 0,
					"upd_ver": "",
					"can_upgrade": 0,
					"connection": 0,
					"is_license_valid": 0
				},
				{
					"name": "fw-branch-01",
					"isGroup": 0,
					"devsn": "FGT60F0000001234",
					"platform_str": "FortiGate-60F",
					"curr_ver": "7.0.1",
					"curr_build": 100,
					"upd_ver": "7.0.5",
					"can_upgrade": 1,
					"connection": 1,
					"is_license_valid": 1
				},
				{
					"name": "fw-dc-01",
					"isGroup": 0,
					"devsn": "FGT3KF0000005678",
					"platform_str": "FortiGate-3000F",
					"curr_ver": "7.0.5",
					"curr_build": 200,
					"upd_ver": "",
					"can_upgrade": 0,
					"connection": 1,
					"is_license_valid": 1
				}
			]`,
		})

		firmware, err := client.ListDeviceFirmware(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(firmware) != 2 {
			t.Fatalf("len = %d, want 2 (group header should be filtered)", len(firmware))
		}

		f := firmware[0]
		if f.Name != "fw-branch-01" {
			t.Errorf("Name = %q", f.Name)
		}
		if f.SerialNumber != "FGT60F0000001234" {
			t.Errorf("SerialNumber = %q", f.SerialNumber)
		}
		if f.Platform != "FortiGate-60F" {
			t.Errorf("Platform = %q", f.Platform)
		}
		if f.CurrentVersion != "7.0.1" {
			t.Errorf("CurrentVersion = %q", f.CurrentVersion)
		}
		if f.CurrentBuild != 100 {
			t.Errorf("CurrentBuild = %d", f.CurrentBuild)
		}
		if f.UpgradeVersion != "7.0.5" {
			t.Errorf("UpgradeVersion = %q", f.UpgradeVersion)
		}
		if !f.CanUpgrade {
			t.Error("CanUpgrade = false, want true")
		}
		if !f.Connected {
			t.Error("Connected = false, want true")
		}
		if !f.LicenseValid {
			t.Error("LicenseValid = false, want true")
		}

		f2 := firmware[1]
		if f2.CanUpgrade {
			t.Error("CanUpgrade = true, want false")
		}
	})
}
