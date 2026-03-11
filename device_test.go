package fortimgr

import (
	"context"
	"testing"
)

func TestListDevices(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListDevices(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom/root/device": `[
				{
					"name": "fw-prod-01",
					"oid": 123,
					"sn": "FGT60F0000000001",
					"platform_str": "FortiGate-60F",
					"os_ver": 7,
					"mr": 2,
					"patch": 5,
					"build": 1517,
					"ha_mode": 0,
					"ha_cluster": 0,
					"conn_status": 1,
					"ip": "10.0.0.1"
				},
				{
					"name": "fw-prod-02",
					"oid": 456,
					"sn": "FGT60F0000000002",
					"platform_str": "FortiGate-60F",
					"os_ver": 7,
					"mr": 4,
					"patch": 1,
					"build": 2093,
					"ha_mode": 1,
					"ha_cluster": 5,
					"conn_status": 0,
					"ip": "10.0.0.2"
				}
			]`,
		})

		devices, err := client.ListDevices(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(devices) != 2 {
			t.Fatalf("len = %d, want 2", len(devices))
		}

		d := devices[0]
		if d.Name != "fw-prod-01" {
			t.Errorf("Name = %q", d.Name)
		}
		if d.DeviceID != "123" {
			t.Errorf("DeviceID = %q", d.DeviceID)
		}
		if d.SerialNumber != "FGT60F0000000001" {
			t.Errorf("SerialNumber = %q", d.SerialNumber)
		}
		if d.Platform != "FortiGate-60F" {
			t.Errorf("Platform = %q", d.Platform)
		}
		if d.Firmware != "7.2.5-b1517" {
			t.Errorf("Firmware = %q", d.Firmware)
		}
		if d.HAMode != "standalone" {
			t.Errorf("HAMode = %q", d.HAMode)
		}
		if d.Status != "online" {
			t.Errorf("Status = %q", d.Status)
		}
		if d.IPAddress != "10.0.0.1" {
			t.Errorf("IPAddress = %q", d.IPAddress)
		}
		if d.ADOM != "root" {
			t.Errorf("ADOM = %q", d.ADOM)
		}

		d2 := devices[1]
		if d2.HAMode != "master" {
			t.Errorf("HAMode = %q, want master", d2.HAMode)
		}
		if d2.HAClusterID != "5" {
			t.Errorf("HAClusterID = %q", d2.HAClusterID)
		}
		if d2.Status != "offline" {
			t.Errorf("Status = %q, want offline", d2.Status)
		}
		if d2.Firmware != "7.4.1-b2093" {
			t.Errorf("Firmware = %q", d2.Firmware)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/dvmdb/adom/test/device": `[]`,
		})
		devices, err := client.ListDevices(context.Background(), "test")
		if err != nil {
			t.Fatal(err)
		}
		if len(devices) != 0 {
			t.Errorf("len = %d, want 0", len(devices))
		}
	})
}
