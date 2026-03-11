package fortimgr

import (
	"context"
	"testing"
)

func TestListIPPools(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListIPPools(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/ippool": `[
				{
					"name": "nat-pool-1",
					"type": "overload",
					"startip": "203.0.113.100",
					"endip": "203.0.113.110",
					"comments": "NAT pool for outbound"
				},
				{
					"name": "nat-pool-2",
					"type": 0,
					"startip": "203.0.113.200",
					"endip": "203.0.113.200",
					"comments": ""
				}
			]`,
		})

		pools, err := client.ListIPPools(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(pools) != 2 {
			t.Fatalf("len = %d, want 2", len(pools))
		}

		p := pools[0]
		if p.Name != "nat-pool-1" {
			t.Errorf("Name = %q", p.Name)
		}
		if p.Type != "overload" {
			t.Errorf("Type = %q", p.Type)
		}
		if p.StartIP != "203.0.113.100" {
			t.Errorf("StartIP = %q", p.StartIP)
		}
		if p.EndIP != "203.0.113.110" {
			t.Errorf("EndIP = %q", p.EndIP)
		}
		if p.Comment != "NAT pool for outbound" {
			t.Errorf("Comment = %q", p.Comment)
		}

		// Enum mapping (int → named string).
		if pools[1].Type != "overload" {
			t.Errorf("Type = %q, want \"overload\"", pools[1].Type)
		}
	})
}
