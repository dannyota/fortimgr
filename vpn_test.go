package fortimgr

import (
	"context"
	"testing"
)

func TestListIPSecPhase1(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListIPSecPhase1(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/vpn.ipsec/phase1-interface": `[
				{
					"name": "vpn-to-branch",
					"interface": "wan1",
					"remote-gw": "203.0.113.1",
					"proposal": "aes256-sha256",
					"dhgrp": "14",
					"mode": 0,
					"type": 0,
					"comments": "Branch office tunnel"
				},
				{
					"name": "vpn-dynamic",
					"interface": ["wan2"],
					"remote-gw": "0.0.0.0",
					"proposal": ["aes128-sha1", "aes256-sha256"],
					"dhgrp": ["14", "5"],
					"mode": 1,
					"type": "dynamic",
					"comments": ""
				}
			]`,
		})

		tunnels, err := client.ListIPSecPhase1(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(tunnels) != 2 {
			t.Fatalf("len = %d, want 2", len(tunnels))
		}

		p := tunnels[0]
		if p.Name != "vpn-to-branch" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.Interface != "wan1" {
			t.Errorf("Interface = %q", p.Interface)
		}
		if p.RemoteGW != "203.0.113.1" {
			t.Errorf("RemoteGW = %q", p.RemoteGW)
		}
		if p.Mode != "main" {
			t.Errorf("Mode = %q, want \"main\"", p.Mode)
		}
		if p.Type != "static" {
			t.Errorf("Type = %q, want \"static\"", p.Type)
		}
		if p.Comments != "Branch office tunnel" {
			t.Errorf("Comments = %q", p.Comments)
		}

		p2 := tunnels[1]
		if p2.Interface != "wan2" {
			t.Errorf("Interface = %q, want \"wan2\"", p2.Interface)
		}
		if p2.Mode != "aggressive" {
			t.Errorf("Mode = %q, want \"aggressive\"", p2.Mode)
		}
		if p2.Type != "dynamic" {
			t.Errorf("Type = %q, want \"dynamic\"", p2.Type)
		}
	})
}

func TestListIPSecPhase2(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListIPSecPhase2(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/vpn.ipsec/phase2-interface": `[
				{
					"name": "vpn-to-branch-p2",
					"phase1name": "vpn-to-branch",
					"proposal": "aes256-sha256",
					"src-subnet": ["10.0.0.0", "255.255.255.0"],
					"dst-subnet": ["192.168.1.0", "255.255.255.0"],
					"comments": "LAN to LAN"
				},
				{
					"name": "vpn-dynamic-p2",
					"phase1name": "vpn-dynamic",
					"proposal": "aes128-sha1",
					"src-subnet": "0.0.0.0/0.0.0.0",
					"dst-subnet": "0.0.0.0/0.0.0.0",
					"comments": ""
				}
			]`,
		})

		tunnels, err := client.ListIPSecPhase2(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(tunnels) != 2 {
			t.Fatalf("len = %d, want 2", len(tunnels))
		}

		p := tunnels[0]
		if p.SrcSubnet != "10.0.0.0/24" {
			t.Errorf("SrcSubnet = %q, want \"10.0.0.0/24\"", p.SrcSubnet)
		}
		if p.DstSubnet != "192.168.1.0/24" {
			t.Errorf("DstSubnet = %q, want \"192.168.1.0/24\"", p.DstSubnet)
		}
		if p.Comments != "LAN to LAN" {
			t.Errorf("Comments = %q", p.Comments)
		}

		if tunnels[1].SrcSubnet != "0.0.0.0/0" {
			t.Errorf("SrcSubnet = %q, want \"0.0.0.0/0\"", tunnels[1].SrcSubnet)
		}
	})
}
