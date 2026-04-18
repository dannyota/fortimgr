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
					"devoid": 123,
					"devsn": "FGT60F0000001234",
					"model": "FGT60F",
					"platform_str": "FortiGate-60F",
					"platform_id": 10,
					"os_type": 0,
					"curr_ver": "7.0.1",
					"curr_build": 100,
					"upd_ver": "7.0.5",
					"upd_ver_key": "0#FortiGate-60F_7.0.5-b123",
					"key_for_download_release": "FortiGate-60F_7.0.5-b123",
					"can_upgrade": 1,
					"connection": 1,
					"is_license_valid": 1,
					"is_model_device": 0,
					"invalid_date": "",
					"upgrade_history": 123,
					"groupName": "7.0.1",
					"status": ["ready"]
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
		if f.DeviceOID != 123 {
			t.Errorf("DeviceOID = %d", f.DeviceOID)
		}
		if f.SerialNumber != "FGT60F0000001234" {
			t.Errorf("SerialNumber = %q", f.SerialNumber)
		}
		if f.Model != "FGT60F" {
			t.Errorf("Model = %q", f.Model)
		}
		if f.Platform != "FortiGate-60F" {
			t.Errorf("Platform = %q", f.Platform)
		}
		if f.PlatformID != 10 || f.OSType != 0 {
			t.Errorf("platform fields = %d/%d", f.PlatformID, f.OSType)
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
		if f.UpgradeVersionKey != "0#FortiGate-60F_7.0.5-b123" || f.DownloadReleaseKey != "FortiGate-60F_7.0.5-b123" {
			t.Errorf("upgrade keys = %q/%q", f.UpgradeVersionKey, f.DownloadReleaseKey)
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
		if f.ModelDevice {
			t.Error("ModelDevice = true, want false")
		}
		if f.UpgradeHistory != 123 || f.GroupName != "7.0.1" {
			t.Errorf("history/group = %d/%q", f.UpgradeHistory, f.GroupName)
		}
		if len(f.Status) != 1 || f.Status[0] != "ready" {
			t.Errorf("Status = %#v", f.Status)
		}

		f2 := firmware[1]
		if f2.CanUpgrade {
			t.Error("CanUpgrade = true, want false")
		}
	})
}

func TestListFirmwareUpgradePaths(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListFirmwareUpgradePaths(context.Background())
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, nil, map[string]string{
			"/gui/adom/dvm/device/firmware:getUpdatePath": `[
				"FGTPlatform=F2K60F|FGTCurrVersion=7.6.5|FGTCurrBuildNum=3651|FGTUpgVersion=7.6.6|FGTUpgBuildNum=3652|BaselineVersion=DISABLE|FGTCurrType=Mature|FGTUpgType=Mature|FGTCurrEOES=20270725\r"
			]`,
		})

		paths, err := client.ListFirmwareUpgradePaths(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) != 1 {
			t.Fatalf("len = %d, want 1", len(paths))
		}
		p := paths[0]
		if p.Platform != "F2K60F" || p.CurrentVersion != "7.6.5" || p.UpgradeVersion != "7.6.6" {
			t.Errorf("versions = %+v", p)
		}
		if p.CurrentBuild != 3651 || p.UpgradeBuild != 3652 {
			t.Errorf("builds = %d/%d", p.CurrentBuild, p.UpgradeBuild)
		}
		if p.BaselineVersion != "DISABLE" || p.CurrentType != "Mature" || p.UpgradeType != "Mature" {
			t.Errorf("metadata = %+v", p)
		}
		if p.CurrentEOES.IsZero() {
			t.Errorf("CurrentEOES not parsed")
		}
	})
}

func TestDevicePSIRT(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.DevicePSIRT(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClientWithProxy(t, map[string]string{
			"/dvmdb/adom": `[{"name":"root","oid":3}]`,
		}, map[string]string{
			"/gui/adoms/3/dvm/psirt:get": `{
				"byIrNumber": {
					"FG-IR-25-001": {
						"ir_number": "FG-IR-25-001",
						"title": "test advisory",
						"summary": "summary",
						"description": "description",
						"risk": 4,
						"threat_severity": "High",
						"cve": ["CVE-2025-0001"],
						"cvss3": {"cvss3_base_score": "8.1", "cvss3_scoring_vector": "AV:N"},
						"products": {"FortiOS": [{"minimum_version": "7.4.0", "maximum_version": "7.4.8", "upgrade_to": "7.4.9"}]},
						"impacted_products": {"FortiOS": [{"major": "7", "minor": "4", "patch": "8"}]}
					}
				},
				"byPlatform": {"FortiOS_7.4.8": ["FG-IR-25-001"]},
				"numDevicesPerRisk": {"4": 2}
			}`,
		})

		report, err := client.DevicePSIRT(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(report.ByIRNumber) != 1 || report.NumDevicesPerRisk["4"] != 2 {
			t.Fatalf("report = %+v", report)
		}
		advisory := report.ByIRNumber["FG-IR-25-001"]
		if advisory.IRNumber != "FG-IR-25-001" || advisory.Risk != 4 || advisory.CVSS3.BaseScore != "8.1" {
			t.Errorf("advisory = %+v", advisory)
		}
		if advisory.Products["FortiOS"][0].UpgradeTo != "7.4.9" {
			t.Errorf("Products = %+v", advisory.Products)
		}
		if advisory.ImpactedProducts["FortiOS"][0].Patch != "8" {
			t.Errorf("ImpactedProducts = %+v", advisory.ImpactedProducts)
		}
	})
}
