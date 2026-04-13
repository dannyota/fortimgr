package fortimgr

import (
	"context"
	"testing"
	"time"
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
					"ip": "10.0.0.1",
					"hostname": "fw-prod-01",
					"conf_status": 1,
					"dev_status": 15,
					"last_checked": 1700000000,
					"last_resync": 1699000000
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
					"ip": "10.0.0.2",
					"hostname": "fw-prod-02",
					"conf_status": 2,
					"dev_status": 4,
					"last_checked": 0,
					"last_resync": 0,
					"ha_slave": [
						{"name": "fw-prod-02",  "role": 1, "sn": "FGT60F0000000002", "status": 1, "conf_status": 1},
						{"name": "fw-prod-02b", "role": 0, "sn": "FGT60F0000000003", "status": 1, "conf_status": 2}
					]
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

		// v1.0.3 new fields.
		if d.Hostname != "fw-prod-01" {
			t.Errorf("Hostname = %q, want fw-prod-01", d.Hostname)
		}
		if d.ConfStatus != "insync" {
			t.Errorf("ConfStatus = %q, want insync", d.ConfStatus)
		}
		if d.DevStatus != "auto_updated" {
			t.Errorf("DevStatus = %q, want auto_updated", d.DevStatus)
		}
		if !d.LastChecked.Equal(time.Unix(1700000000, 0).UTC()) {
			t.Errorf("LastChecked = %v, want 1700000000 UTC", d.LastChecked)
		}
		if !d.LastResync.Equal(time.Unix(1699000000, 0).UTC()) {
			t.Errorf("LastResync = %v", d.LastResync)
		}
		if d.HARole != "" {
			t.Errorf("HARole = %q, want empty (standalone has no ha_slave)", d.HARole)
		}
		if d.HAMembers != nil {
			t.Errorf("HAMembers should be nil for standalone device, got %+v", d.HAMembers)
		}

		d2 := devices[1]
		// HAMode stays on the legacy ha_mode int mapping for backwards compat.
		if d2.HAMode != "master" {
			t.Errorf("HAMode = %q, want master (legacy mapping)", d2.HAMode)
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

		// v1.0.3 HA role derivation — matches top-level name against ha_slave[].
		if d2.HARole != "master" {
			t.Errorf("HARole = %q, want master (role=1 in matching ha_slave entry)", d2.HARole)
		}
		if d2.ConfStatus != "modified" {
			t.Errorf("ConfStatus = %q, want modified", d2.ConfStatus)
		}
		if d2.DevStatus != "installed" {
			t.Errorf("DevStatus = %q, want installed", d2.DevStatus)
		}
		// Zero timestamps map to zero time.Time, not 1970.
		if !d2.LastChecked.IsZero() {
			t.Errorf("LastChecked should be zero when last_checked=0, got %v", d2.LastChecked)
		}
		if !d2.LastResync.IsZero() {
			t.Errorf("LastResync should be zero when last_resync=0, got %v", d2.LastResync)
		}

		// HAMembers — both primary and standby surface here.
		if len(d2.HAMembers) != 2 {
			t.Fatalf("HAMembers len = %d, want 2", len(d2.HAMembers))
		}
		m0 := d2.HAMembers[0]
		if m0.Name != "fw-prod-02" || m0.SerialNumber != "FGT60F0000000002" {
			t.Errorf("HAMembers[0] name/sn = %q/%q", m0.Name, m0.SerialNumber)
		}
		if m0.Role != "master" || m0.Status != "online" || m0.ConfStatus != "insync" {
			t.Errorf("HAMembers[0] role/status/conf = %q/%q/%q, want master/online/insync",
				m0.Role, m0.Status, m0.ConfStatus)
		}
		m1 := d2.HAMembers[1]
		if m1.Name != "fw-prod-02b" || m1.SerialNumber != "FGT60F0000000003" {
			t.Errorf("HAMembers[1] name/sn = %q/%q", m1.Name, m1.SerialNumber)
		}
		if m1.Role != "slave" {
			t.Errorf("HAMembers[1] role = %q, want slave (standby)", m1.Role)
		}
		if m1.ConfStatus != "modified" {
			t.Errorf("HAMembers[1] conf_status = %q, want modified", m1.ConfStatus)
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
