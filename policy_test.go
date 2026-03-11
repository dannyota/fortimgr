package fortimgr

import (
	"context"
	"testing"
)

func TestListPolicyPackages(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListPolicyPackages(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("filters folders", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/pkg/adom/root": `[
				{
					"name": "default",
					"type": "pkg",
					"scope member": [{"name": "fw-01"}, {"name": "fw-02"}]
				},
				{
					"name": "folder1",
					"type": "folder"
				},
				{
					"name": "dmz-policy",
					"type": "pkg",
					"scope member": []
				}
			]`,
		})

		pkgs, err := client.ListPolicyPackages(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(pkgs) != 2 {
			t.Fatalf("len = %d, want 2 (folder filtered)", len(pkgs))
		}
		if pkgs[0].Name != "default" {
			t.Errorf("Name = %q", pkgs[0].Name)
		}
		if len(pkgs[0].Scope) != 2 {
			t.Errorf("Scope len = %d, want 2", len(pkgs[0].Scope))
		}
		if pkgs[0].ADOM != "root" {
			t.Errorf("ADOM = %q", pkgs[0].ADOM)
		}
		if pkgs[1].Name != "dmz-policy" {
			t.Errorf("Name = %q", pkgs[1].Name)
		}
	})
}

func TestListPolicies(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListPolicies(context.Background(), "root", "default")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/pkg/default/firewall/policy": `[
				{
					"policyid": 1,
					"name": "Allow-Web",
					"srcintf": ["port1"],
					"dstintf": ["port2"],
					"srcaddr": ["all"],
					"dstaddr": ["web-servers"],
					"service": ["HTTP", "HTTPS"],
					"action": "accept",
					"schedule": "always",
					"nat": "enable",
					"status": "enable",
					"logtraffic": "all",
					"comments": "Allow web traffic"
				},
				{
					"policyid": 2,
					"name": "Deny-All",
					"srcintf": "any",
					"dstintf": "any",
					"srcaddr": "all",
					"dstaddr": "all",
					"service": "ALL",
					"action": 0,
					"schedule": "always",
					"nat": 0,
					"status": 1,
					"logtraffic": "utm",
					"comments": ""
				}
			]`,
		})

		policies, err := client.ListPolicies(context.Background(), "root", "default")
		if err != nil {
			t.Fatal(err)
		}
		if len(policies) != 2 {
			t.Fatalf("len = %d, want 2", len(policies))
		}

		p := policies[0]
		if p.PolicyID != 1 {
			t.Errorf("PolicyID = %d", p.PolicyID)
		}
		if p.Name != "Allow-Web" {
			t.Errorf("Name = %q", p.Name)
		}
		if len(p.SrcIntf) != 1 || p.SrcIntf[0] != "port1" {
			t.Errorf("SrcIntf = %v", p.SrcIntf)
		}
		if len(p.DstIntf) != 1 || p.DstIntf[0] != "port2" {
			t.Errorf("DstIntf = %v", p.DstIntf)
		}
		if len(p.SrcAddr) != 1 || p.SrcAddr[0] != "all" {
			t.Errorf("SrcAddr = %v", p.SrcAddr)
		}
		if len(p.DstAddr) != 1 || p.DstAddr[0] != "web-servers" {
			t.Errorf("DstAddr = %v", p.DstAddr)
		}
		if len(p.Service) != 2 {
			t.Errorf("Service = %v", p.Service)
		}
		if p.Action != "accept" {
			t.Errorf("Action = %q", p.Action)
		}
		if p.Schedule != "always" {
			t.Errorf("Schedule = %q", p.Schedule)
		}
		if p.NAT != "enable" {
			t.Errorf("NAT = %q", p.NAT)
		}
		if p.Status != "enable" {
			t.Errorf("Status = %q", p.Status)
		}
		if p.LogTraffic != "all" {
			t.Errorf("LogTraffic = %q", p.LogTraffic)
		}
		if p.Comments != "Allow web traffic" {
			t.Errorf("Comments = %q", p.Comments)
		}

		// Test enum mapping (int → named string).
		p2 := policies[1]
		if p2.Action != "deny" {
			t.Errorf("Action = %q, want \"deny\"", p2.Action)
		}
		if len(p2.SrcIntf) != 1 || p2.SrcIntf[0] != "any" {
			t.Errorf("SrcIntf = %v (string → []string)", p2.SrcIntf)
		}
		if p2.NAT != "disable" {
			t.Errorf("NAT = %q, want \"disable\"", p2.NAT)
		}
		if p2.Status != "enable" {
			t.Errorf("Status = %q, want \"enable\"", p2.Status)
		}
		if p2.LogTraffic != "utm" {
			t.Errorf("LogTraffic = %q, want \"utm\"", p2.LogTraffic)
		}
		if p2.Comments != "" {
			t.Errorf("Comments = %q, want empty", p2.Comments)
		}
	})
}
