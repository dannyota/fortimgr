package fortimgr

import (
	"context"
	"testing"
)

func TestSystemStatus(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.SystemStatus(context.Background())
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/sys/config": `{
				"hostname": "fmg-lab",
				"sn": "FMG-VM0000000001",
				"fmgversion": "v7.2.0-build1000 230101 (GA)",
				"build_number": 1000,
				"ha_mode": 0,
				"platform-id": "FortiManager-VM64"
			}`,
		})

		status, err := client.SystemStatus(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if status.Hostname != "fmg-lab" {
			t.Errorf("Hostname = %q", status.Hostname)
		}
		if status.SerialNumber != "FMG-VM0000000001" {
			t.Errorf("SerialNumber = %q", status.SerialNumber)
		}
		if status.Version != "v7.2.0-build1000 230101 (GA)" {
			t.Errorf("Version = %q", status.Version)
		}
		if status.Platform != "FortiManager-VM64" {
			t.Errorf("Platform = %q", status.Platform)
		}
		if status.Build != 1000 {
			t.Errorf("Build = %d, want 1000", status.Build)
		}
		if status.HAMode != "standalone" {
			t.Errorf("HAMode = %q, want \"standalone\"", status.HAMode)
		}
	})
}
