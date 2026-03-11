package fortimgr

import (
	"context"
	"testing"
)

func TestListStaticRoutes(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListStaticRoutes(context.Background(), "fw-01", "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/device/fw-01/vdom/root/router/static": `[
				{
					"seq-num": 1,
					"dst": ["0.0.0.0", "0.0.0.0"],
					"gateway": "10.0.0.1",
					"device": "wan1",
					"distance": 10,
					"priority": 0,
					"status": 1,
					"comment": "Default route"
				},
				{
					"seq-num": 2,
					"dst": ["192.168.10.0", "255.255.255.0"],
					"gateway": "10.0.1.1",
					"device": "port2",
					"distance": 20,
					"priority": 5,
					"status": 0,
					"comment": ""
				}
			]`,
		})

		routes, err := client.ListStaticRoutes(context.Background(), "fw-01", "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(routes) != 2 {
			t.Fatalf("len = %d, want 2", len(routes))
		}

		r := routes[0]
		if r.SeqNum != 1 {
			t.Errorf("SeqNum = %d", r.SeqNum)
		}
		if r.Dst != "0.0.0.0/0" {
			t.Errorf("Dst = %q, want \"0.0.0.0/0\"", r.Dst)
		}
		if r.Gateway != "10.0.0.1" {
			t.Errorf("Gateway = %q", r.Gateway)
		}
		if r.Device != "wan1" {
			t.Errorf("Device = %q", r.Device)
		}
		if r.Distance != 10 {
			t.Errorf("Distance = %d", r.Distance)
		}
		if r.Status != "enable" {
			t.Errorf("Status = %q, want \"enable\"", r.Status)
		}
		if r.Comment != "Default route" {
			t.Errorf("Comment = %q", r.Comment)
		}

		r2 := routes[1]
		if r2.Dst != "192.168.10.0/24" {
			t.Errorf("Dst = %q, want \"192.168.10.0/24\"", r2.Dst)
		}
		if r2.Status != "disable" {
			t.Errorf("Status = %q, want \"disable\"", r2.Status)
		}
		if r2.Priority != 5 {
			t.Errorf("Priority = %d, want 5", r2.Priority)
		}
	})
}
