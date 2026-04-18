package fortimgr

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeviceSummary(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.DeviceSummary(context.Background(), "root", "fw-01")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{})
		_, err := client.DeviceSummary(context.Background(), "root", "bad/device")
		if !errors.Is(err, ErrInvalidName) {
			t.Errorf("err = %v, want ErrInvalidName", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/adom/dvm/device/summary:get": `{
				"devices": [
					{
						"name": "fw-01",
						"conf_status": 1,
						"install": {
							"total_revision": 42,
							"last_install_rev": 41,
							"last_install_time": 1700000000,
							"installed_by": "admin"
						},
						"ha_mode": 1,
						"ha_upgrade_mode": 2,
						"ha_cluster_name": "cluster-a",
						"ha_cluster_id": 9,
						"ha": [
							{"name": "fw-01", "sn": "FGT0001", "role": 1, "sync_status": 1},
							{"name": "fw-02", "serial": "FGT0002", "role": 0, "sync_status": 2}
						]
					}
				]
			}`,
		})

		s, err := client.DeviceSummary(context.Background(), "root", "fw-01")
		if err != nil {
			t.Fatal(err)
		}
		if s.ADOM != "root" || s.Device != "fw-01" {
			t.Errorf("identity = %+v", s)
		}
		if s.ConfigStatus != "insync" {
			t.Errorf("ConfigStatus = %q, want insync", s.ConfigStatus)
		}
		if s.TotalRevisions != 42 || s.LastInstalledRevision != 41 {
			t.Errorf("revision fields = %d/%d", s.TotalRevisions, s.LastInstalledRevision)
		}
		if want := time.Unix(1700000000, 0).UTC(); !s.LastInstallTime.Equal(want) {
			t.Errorf("LastInstallTime = %s, want %s", s.LastInstallTime, want)
		}
		if s.InstalledBy != "admin" {
			t.Errorf("InstalledBy = %q", s.InstalledBy)
		}
		if s.HAMode != "active-passive" || s.HAUpgradeMode != "uninterruptible" {
			t.Errorf("HA modes = %q/%q", s.HAMode, s.HAUpgradeMode)
		}
		if s.HAClusterName != "cluster-a" || s.HAClusterID != 9 {
			t.Errorf("cluster = %q/%d", s.HAClusterName, s.HAClusterID)
		}
		if len(s.HAMembers) != 2 {
			t.Fatalf("HAMembers len = %d, want 2", len(s.HAMembers))
		}
		if s.HAMembers[0].Role != "master" || s.HAMembers[1].Role != "slave" {
			t.Errorf("member roles = %+v", s.HAMembers)
		}
	})

	t.Run("live dashboard shape", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/adom/dvm/device/summary:get": `{
				"sysConfig": {
					"syncStatus": {"value": "Synchronized"},
					"revision": {"value": 42},
					"installTracking": {
						"lastInstallation": {
							"revision": 41,
							"value": "Revision-41 (2026-04-17 18:07:36) Installed By: admin"
						}
					}
				},
				"sysInfo": {
					"hostName": {"value": "fw-host"},
					"sn": {"value": "FGT0001"},
					"firmware": {"value": "v7.4.7"},
					"haName": {"value": "cluster-a"},
					"haStatus": {"value": "Standalone"},
					"haMember": {
						"value": {
							"records": [
								{"oid": 1, "name": "fw-host", "sn": "FGT0001", "role": "Primary", "status": 1}
							]
						}
					}
				}
			}`,
		})

		s, err := client.DeviceSummary(context.Background(), "root", "fw-01")
		if err != nil {
			t.Fatal(err)
		}
		if s.Hostname != "fw-host" || s.SerialNumber != "FGT0001" || s.Firmware != "v7.4.7" {
			t.Errorf("system fields = %+v", s)
		}
		if s.ConfigStatus != "Synchronized" || s.TotalRevisions != 42 || s.LastInstalledRevision != 41 {
			t.Errorf("config fields = %+v", s)
		}
		if s.LastInstallation == "" || s.InstalledBy != "admin" {
			t.Errorf("install fields = %+v", s)
		}
		if s.LastInstallTime.IsZero() {
			t.Errorf("LastInstallTime was not parsed")
		}
		if s.HAMode != "Standalone" || s.HAClusterName != "cluster-a" {
			t.Errorf("HA fields = %+v", s)
		}
		if len(s.HAMembers) != 1 || s.HAMembers[0].OID != 1 || s.HAMembers[0].Status != 1 {
			t.Errorf("HAMembers = %+v", s.HAMembers)
		}
	})

	t.Run("no matching device returns empty summary", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/adom/dvm/device/summary:get": `{"devices":[]}`,
		})

		s, err := client.DeviceSummary(context.Background(), "root", "fw-01")
		if err != nil {
			t.Fatal(err)
		}
		if s.ADOM != "root" || s.Device != "fw-01" || s.TotalRevisions != 0 {
			t.Errorf("summary = %+v", s)
		}
	})
}
