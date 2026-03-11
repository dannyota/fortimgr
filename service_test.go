package fortimgr

import (
	"context"
	"testing"
)

func TestListServices(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListServices(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/service/custom": `[
				{
					"name": "HTTP",
					"protocol": "TCP/UDP/SCTP",
					"tcp-portrange": "80",
					"udp-portrange": "",
					"comment": "Web traffic"
				},
				{
					"name": "Custom-App",
					"protocol": 6,
					"tcp-portrange": ["8080", "8443"],
					"udp-portrange": "9000-9100",
					"comment": ""
				}
			]`,
		})

		services, err := client.ListServices(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(services) != 2 {
			t.Fatalf("len = %d, want 2", len(services))
		}

		s := services[0]
		if s.Name != "HTTP" {
			t.Errorf("Name = %q", s.Name)
		}
		if s.Protocol != "TCP/UDP/SCTP" {
			t.Errorf("Protocol = %q", s.Protocol)
		}
		if s.TCPRange != "80" {
			t.Errorf("TCPRange = %q", s.TCPRange)
		}
		if s.Comment != "Web traffic" {
			t.Errorf("Comment = %q", s.Comment)
		}

		// Test enum mapping (protocol int → named string, port range as array).
		s2 := services[1]
		if s2.Protocol != "ICMP6" {
			t.Errorf("Protocol = %q, want \"ICMP6\"", s2.Protocol)
		}
		if s2.TCPRange != "8080" {
			t.Errorf("TCPRange = %q, want \"8080\" (first of array)", s2.TCPRange)
		}
		if s2.UDPRange != "9000-9100" {
			t.Errorf("UDPRange = %q", s2.UDPRange)
		}
	})
}

func TestListServiceGroups(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListServiceGroups(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/service/group": `[
				{
					"name": "Web-Services",
					"member": ["HTTP", "HTTPS", "DNS"],
					"comment": "Standard web services"
				}
			]`,
		})

		groups, err := client.ListServiceGroups(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 1 {
			t.Fatalf("len = %d, want 1", len(groups))
		}

		g := groups[0]
		if g.Name != "Web-Services" {
			t.Errorf("Name = %q", g.Name)
		}
		if len(g.Members) != 3 {
			t.Errorf("Members = %v", g.Members)
		}
		if g.Comment != "Standard web services" {
			t.Errorf("Comment = %q", g.Comment)
		}
	})
}
