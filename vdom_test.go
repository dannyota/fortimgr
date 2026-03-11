package fortimgr

import (
	"context"
	"testing"
)

func TestListVDOMs(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListVDOMs(context.Background(), "fw-01")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/dvmdb/device/fw-01/vdom": `[
				{
					"name": "root",
					"status": 1,
					"opmode": 0
				},
				{
					"name": "dmz",
					"status": "enable",
					"opmode": 1
				}
			]`,
		})

		vdoms, err := client.ListVDOMs(context.Background(), "fw-01")
		if err != nil {
			t.Fatal(err)
		}
		if len(vdoms) != 2 {
			t.Fatalf("len = %d, want 2", len(vdoms))
		}

		v := vdoms[0]
		if v.Name != "root" {
			t.Errorf("Name = %q", v.Name)
		}
		if v.Status != "enable" {
			t.Errorf("Status = %q, want \"enable\"", v.Status)
		}
		if v.OpMode != "nat" {
			t.Errorf("OpMode = %q, want \"nat\"", v.OpMode)
		}

		v2 := vdoms[1]
		if v2.Name != "dmz" {
			t.Errorf("Name = %q", v2.Name)
		}
		if v2.OpMode != "transparent" {
			t.Errorf("OpMode = %q, want \"transparent\"", v2.OpMode)
		}
	})
}
