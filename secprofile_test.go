package fortimgr

import (
	"context"
	"testing"
)

func TestListAntivirusProfiles(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListAntivirusProfiles(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/antivirus/profile": `[
				{
					"name": "default",
					"comment": "Default antivirus profile",
					"scan-mode": "full",
					"feature-set": "flow",
					"av-block-log": 1,
					"av-virus-log": "enable",
					"extended-log": 0,
					"analytics-db": 1,
					"mobile-malware-db": 0
				},
				{
					"name": "wifi-av",
					"comment": "",
					"scan-mode": 0,
					"feature-set": 1,
					"av-block-log": "disable",
					"av-virus-log": 1,
					"extended-log": "disable",
					"analytics-db": "disable",
					"mobile-malware-db": "enable"
				}
			]`,
		})

		profiles, err := client.ListAntivirusProfiles(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(profiles) != 2 {
			t.Fatalf("len = %d, want 2", len(profiles))
		}

		p := profiles[0]
		if p.Name != "default" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.ScanMode != "full" {
			t.Errorf("ScanMode = %q", p.ScanMode)
		}
		if p.FeatureSet != "flow" {
			t.Errorf("FeatureSet = %q, want \"flow\"", p.FeatureSet)
		}
		if p.AVBlockLog != "enable" {
			t.Errorf("AVBlockLog = %q, want \"enable\"", p.AVBlockLog)
		}
		if p.AVVirusLog != "enable" {
			t.Errorf("AVVirusLog = %q, want \"enable\"", p.AVVirusLog)
		}
		if p.ExtendedLog != "disable" {
			t.Errorf("ExtendedLog = %q, want \"disable\"", p.ExtendedLog)
		}
		if p.AnalyticsDB != "enable" {
			t.Errorf("AnalyticsDB = %q, want \"enable\"", p.AnalyticsDB)
		}
		if p.MobileMalware != "disable" {
			t.Errorf("MobileMalware = %q, want \"disable\"", p.MobileMalware)
		}

		p2 := profiles[1]
		if p2.ScanMode != "default" {
			t.Errorf("ScanMode = %q, want \"default\"", p2.ScanMode)
		}
		if p2.FeatureSet != "proxy" {
			t.Errorf("FeatureSet = %q, want \"proxy\"", p2.FeatureSet)
		}
		if p2.MobileMalware != "enable" {
			t.Errorf("MobileMalware = %q, want \"enable\"", p2.MobileMalware)
		}
	})
}

func TestListIPSSensors(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListIPSSensors(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/ips/sensor": `[
				{
					"name": "all_default",
					"comment": "Prevent critical attacks",
					"extended-log": 1,
					"block-malicious-url": "enable",
					"scan-botnet-connections": 2
				},
				{
					"name": "protect_http_server",
					"comment": "",
					"extended-log": "disable",
					"block-malicious-url": 0,
					"scan-botnet-connections": "disable"
				}
			]`,
		})

		sensors, err := client.ListIPSSensors(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(sensors) != 2 {
			t.Fatalf("len = %d, want 2", len(sensors))
		}

		s := sensors[0]
		if s.Name != "all_default" {
			t.Errorf("Name = %q", s.Name)
		}
		if s.ExtendedLog != "enable" {
			t.Errorf("ExtendedLog = %q, want \"enable\"", s.ExtendedLog)
		}
		if s.BlockMaliciousURL != "enable" {
			t.Errorf("BlockMaliciousURL = %q, want \"enable\"", s.BlockMaliciousURL)
		}
		if s.ScanBotnetConnections != "monitor" {
			t.Errorf("ScanBotnetConnections = %q, want \"monitor\"", s.ScanBotnetConnections)
		}

		s2 := sensors[1]
		if s2.ExtendedLog != "disable" {
			t.Errorf("ExtendedLog = %q, want \"disable\"", s2.ExtendedLog)
		}
		if s2.BlockMaliciousURL != "disable" {
			t.Errorf("BlockMaliciousURL = %q, want \"disable\"", s2.BlockMaliciousURL)
		}
		if s2.ScanBotnetConnections != "disable" {
			t.Errorf("ScanBotnetConnections = %q, want \"disable\"", s2.ScanBotnetConnections)
		}
	})
}

func TestListWebFilterProfiles(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListWebFilterProfiles(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/webfilter/profile": `[
				{
					"name": "default",
					"comment": "Default web filter",
					"feature-set": "flow",
					"inspection-mode": 1,
					"log-all-url": 0,
					"web-content-log": 1,
					"web-ftgd-err-log": "enable",
					"extended-log": "disable"
				},
				{
					"name": "monitor-all",
					"comment": "Monitor without blocking",
					"feature-set": 0,
					"inspection-mode": "proxy",
					"log-all-url": "enable",
					"web-content-log": "enable",
					"web-ftgd-err-log": 1,
					"extended-log": 1
				}
			]`,
		})

		profiles, err := client.ListWebFilterProfiles(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(profiles) != 2 {
			t.Fatalf("len = %d, want 2", len(profiles))
		}

		p := profiles[0]
		if p.Name != "default" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.FeatureSet != "flow" {
			t.Errorf("FeatureSet = %q, want \"flow\"", p.FeatureSet)
		}
		if p.InspectionMode != "flow-based" {
			t.Errorf("InspectionMode = %q, want \"flow-based\"", p.InspectionMode)
		}
		if p.LogAllURL != "disable" {
			t.Errorf("LogAllURL = %q, want \"disable\"", p.LogAllURL)
		}
		if p.WebContentLog != "enable" {
			t.Errorf("WebContentLog = %q, want \"enable\"", p.WebContentLog)
		}

		p2 := profiles[1]
		if p2.InspectionMode != "proxy" {
			t.Errorf("InspectionMode = %q, want \"proxy\"", p2.InspectionMode)
		}
		if p2.LogAllURL != "enable" {
			t.Errorf("LogAllURL = %q, want \"enable\"", p2.LogAllURL)
		}
		if p2.ExtendedLog != "enable" {
			t.Errorf("ExtendedLog = %q, want \"enable\"", p2.ExtendedLog)
		}
	})
}

func TestListAppControlProfiles(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListAppControlProfiles(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/application/list": `[
				{
					"name": "default",
					"comment": "Monitor all applications",
					"extended-log": 0,
					"deep-app-inspection": 1,
					"enforce-default-app-port": "enable",
					"unknown-application-action": 0,
					"unknown-application-log": 1,
					"other-application-action": "block",
					"other-application-log": "enable"
				},
				{
					"name": "block-p2p",
					"comment": "Block peer-to-peer",
					"extended-log": "enable",
					"deep-app-inspection": "enable",
					"enforce-default-app-port": 0,
					"unknown-application-action": 1,
					"unknown-application-log": "disable",
					"other-application-action": 1,
					"other-application-log": 0
				}
			]`,
		})

		profiles, err := client.ListAppControlProfiles(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(profiles) != 2 {
			t.Fatalf("len = %d, want 2", len(profiles))
		}

		p := profiles[0]
		if p.Name != "default" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.DeepAppInspection != "enable" {
			t.Errorf("DeepAppInspection = %q, want \"enable\"", p.DeepAppInspection)
		}
		if p.UnknownApplicationAction != "pass" {
			t.Errorf("UnknownApplicationAction = %q, want \"pass\"", p.UnknownApplicationAction)
		}
		if p.UnknownApplicationLog != "enable" {
			t.Errorf("UnknownApplicationLog = %q, want \"enable\"", p.UnknownApplicationLog)
		}
		if p.OtherApplicationAction != "block" {
			t.Errorf("OtherApplicationAction = %q, want \"block\"", p.OtherApplicationAction)
		}

		p2 := profiles[1]
		if p2.ExtendedLog != "enable" {
			t.Errorf("ExtendedLog = %q, want \"enable\"", p2.ExtendedLog)
		}
		if p2.UnknownApplicationAction != "block" {
			t.Errorf("UnknownApplicationAction = %q, want \"block\"", p2.UnknownApplicationAction)
		}
		if p2.EnforceDefaultAppPort != "disable" {
			t.Errorf("EnforceDefaultAppPort = %q, want \"disable\"", p2.EnforceDefaultAppPort)
		}
	})
}

func TestListSSLSSHProfiles(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListSSLSSHProfiles(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/ssl-ssh-profile": `[
				{
					"name": "certificate-inspection",
					"comment": "Read-only inspection",
					"server-cert-mode": 0,
					"mapi-over-https": 1,
					"rpc-over-https": "enable",
					"ssl-anomaly-log": 1,
					"ssl-exemption-log": "enable",
					"supported-alpn": 3
				},
				{
					"name": "deep-inspection",
					"comment": "",
					"server-cert-mode": "replace",
					"mapi-over-https": "disable",
					"rpc-over-https": 0,
					"ssl-anomaly-log": "enable",
					"ssl-exemption-log": 0,
					"supported-alpn": "none"
				}
			]`,
		})

		profiles, err := client.ListSSLSSHProfiles(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(profiles) != 2 {
			t.Fatalf("len = %d, want 2", len(profiles))
		}

		p := profiles[0]
		if p.Name != "certificate-inspection" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.ServerCertMode != "re-sign" {
			t.Errorf("ServerCertMode = %q, want \"re-sign\"", p.ServerCertMode)
		}
		if p.MAPIOverHTTPS != "enable" {
			t.Errorf("MAPIOverHTTPS = %q, want \"enable\"", p.MAPIOverHTTPS)
		}
		if p.RPCOverHTTPS != "enable" {
			t.Errorf("RPCOverHTTPS = %q, want \"enable\"", p.RPCOverHTTPS)
		}
		if p.SSLAnomalyLog != "enable" {
			t.Errorf("SSLAnomalyLog = %q, want \"enable\"", p.SSLAnomalyLog)
		}
		if p.SSLExemptionLog != "enable" {
			t.Errorf("SSLExemptionLog = %q, want \"enable\"", p.SSLExemptionLog)
		}
		if p.SupportedALPN != "all" {
			t.Errorf("SupportedALPN = %q, want \"all\"", p.SupportedALPN)
		}

		p2 := profiles[1]
		if p2.ServerCertMode != "replace" {
			t.Errorf("ServerCertMode = %q, want \"replace\"", p2.ServerCertMode)
		}
		if p2.MAPIOverHTTPS != "disable" {
			t.Errorf("MAPIOverHTTPS = %q, want \"disable\"", p2.MAPIOverHTTPS)
		}
		if p2.SSLExemptionLog != "disable" {
			t.Errorf("SSLExemptionLog = %q, want \"disable\"", p2.SSLExemptionLog)
		}
		if p2.SupportedALPN != "none" {
			t.Errorf("SupportedALPN = %q, want \"none\"", p2.SupportedALPN)
		}
	})
}
