package fortimgr

import (
	"context"
	"testing"
)

func TestDeviceDNS(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/device/fw-01/global/system/dns": `{
				"oid": 1,
				"primary": "8.8.8.8",
				"secondary": "1.1.1.1",
				"ip6-primary": "2001:4860:4860::8888",
				"protocol": 2,
				"server-hostname": ["dns.example"],
				"domain": ["example.com"],
				"interface": ["wan1"],
				"source-ip": "192.0.2.10",
				"retry": 3,
				"timeout": 5,
				"dns-cache-limit": 5000,
				"cache-notfound-responses": 1,
				"log": 0
			}`,
		})

		dns, err := client.DeviceDNS(context.Background(), "fw-01")
		if err != nil {
			t.Fatal(err)
		}
		if dns.Primary != "8.8.8.8" || dns.Secondary != "1.1.1.1" || dns.IPv6Primary != "2001:4860:4860::8888" {
			t.Errorf("DNS servers = %+v", dns)
		}
		if dns.CacheNotFoundResponses != "enable" || dns.Log != "disable" {
			t.Errorf("enum fields = %+v", dns)
		}
		if len(dns.Domain) != 1 || dns.Domain[0] != "example.com" {
			t.Errorf("Domain = %#v", dns.Domain)
		}
	})
}

func TestListDeviceDDNS(t *testing.T) {
	client := newTestClient(t, map[string]string{
		"/pm/config/device/fw-01/global/system/ddns": `[
			{"name":"ddns1","ddns-domain":"fw.example.com","monitor-interface":"wan1","server":1,"status":1,"username":"user"}
		]`,
	})
	items, err := client.ListDeviceDDNS(context.Background(), "fw-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].Name != "ddns1" || items[0].Status != "enable" || items[0].Username != "user" {
		t.Errorf("DDNS = %+v", items[0])
	}
}

func TestDeviceIPAM(t *testing.T) {
	client := newTestClient(t, map[string]string{
		"/pm/config/device/fw-01/vdom/root/system/ipam": `{
			"oid": 2,
			"status": 1,
			"server-type": 0,
			"automatic-conflict-resolution": 1,
			"manage-lan-addresses": 1,
			"manage-lan-extension-addresses": 0,
			"manage-ssid-addresses": 1,
			"require-subnet-size-match": 0
		}`,
	})
	ipam, err := client.DeviceIPAM(context.Background(), "fw-01", "root")
	if err != nil {
		t.Fatal(err)
	}
	if ipam.Status != "enable" || ipam.AutomaticConflictResolution != "enable" || ipam.ManageLANExtensionAddresses != "disable" {
		t.Errorf("IPAM = %+v", ipam)
	}
}

func TestSDWANSettings(t *testing.T) {
	client := newTestClient(t, map[string]string{
		"/pm/config/device/fw-01/vdom/root/system/sdwan": `{
			"oid": 3,
			"status": 1,
			"load-balance-mode": 2,
			"fail-detect": 1,
			"app-perf-log-period": 60,
			"duplication-max-num": 2,
			"neighbor-hold-down": 0,
			"fail-alert-interfaces": ["wan1"],
			"health-check": [{
				"name": "hc1",
				"oid": 10,
				"protocol": 1,
				"server": ["8.8.8.8"],
				"members": [1],
				"interval": 500,
				"failtime": 5,
				"recoverytime": 5,
				"system-dns": 1,
				"update-static-route": 1,
				"sla": [{"id": 1, "oid": 11, "latency-threshold": 50, "jitter-threshold": 10, "packetloss-threshold": 5, "link-cost-factor": 2}]
			}],
			"zone": [{"name": "virtual-wan-link", "oid": 20, "minimum-sla-meet-members": 1, "service-sla-tie-break": 1, "advpn-select": 0}]
		}`,
	})
	settings, err := client.SDWANSettings(context.Background(), "fw-01", "root")
	if err != nil {
		t.Fatal(err)
	}
	if settings.Status != "enable" || settings.FailDetect != "enable" || settings.NeighborHoldDown != "disable" {
		t.Errorf("settings = %+v", settings)
	}
	if len(settings.HealthChecks) != 1 || settings.HealthChecks[0].SystemDNS != "enable" || len(settings.HealthChecks[0].SLA) != 1 {
		t.Errorf("health checks = %+v", settings.HealthChecks)
	}
	if len(settings.Zones) != 1 || settings.Zones[0].Name != "virtual-wan-link" {
		t.Errorf("zones = %+v", settings.Zones)
	}
}

func TestSDWANLists(t *testing.T) {
	client := newTestClient(t, map[string]string{
		"/pm/config/device/fw-01/vdom/root/system/sdwan/members": `[
			{"seq-num":1,"interface":"wan1","gateway":"192.0.2.1","gateway6":"2001:db8::1","zone":"virtual-wan-link","status":1,"cost":10,"priority":1,"comment":"primary"}
		]`,
		"/pm/config/device/fw-01/vdom/root/system/sdwan/service": `[
			{"id":1,"name":"svc1","mode":1,"status":1,"addr-mode":0,"input-device":["lan"],"src":["all"],"dst":["all"],"internet-service":0,"priority-members":[1]}
		]`,
		"/pm/config/device/fw-01/vdom/root/system/sdwan/duplication": `[
			{"id":1,"name":"dup1","service-id":1,"packet-duplication":1,"fields":["src-ip"]}
		]`,
	})
	members, err := client.ListSDWANMembers(context.Background(), "fw-01", "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 1 || members[0].Status != "enable" || members[0].Gateway6 != "2001:db8::1" {
		t.Errorf("members = %+v", members)
	}
	services, err := client.ListSDWANServices(context.Background(), "fw-01", "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].InternetService != "disable" || services[0].PriorityMembers[0] != "1" {
		t.Errorf("services = %+v", services)
	}
	duplication, err := client.ListSDWANDuplication(context.Background(), "fw-01", "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(duplication) != 1 || duplication[0].PacketDuplication != "1" || duplication[0].Fields[0] != "src-ip" {
		t.Errorf("duplication = %+v", duplication)
	}
}
