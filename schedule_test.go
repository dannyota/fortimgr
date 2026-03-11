package fortimgr

import (
	"context"
	"testing"
)

func TestListSchedulesRecurring(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListSchedulesRecurring(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/schedule/recurring": `[
				{
					"name": "weekday-business",
					"day": 62,
					"start": "08:00",
					"end": "18:00",
					"color": 2
				},
				{
					"name": "always",
					"day": 127,
					"start": "00:00",
					"end": "00:00",
					"color": 0
				},
				{
					"name": "none",
					"day": 0,
					"start": "00:00",
					"end": "00:00",
					"color": 0
				}
			]`,
		})

		schedules, err := client.ListSchedulesRecurring(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(schedules) != 3 {
			t.Fatalf("len = %d, want 3", len(schedules))
		}

		// Bitmask 62 = monday through friday.
		s := schedules[0]
		if s.Name != "weekday-business" {
			t.Errorf("Name = %q", s.Name)
		}
		if s.Type != "recurring" {
			t.Errorf("Type = %q", s.Type)
		}
		if s.Day != "monday tuesday wednesday thursday friday" {
			t.Errorf("Day = %q", s.Day)
		}
		if s.Start != "08:00" {
			t.Errorf("Start = %q", s.Start)
		}
		if s.End != "18:00" {
			t.Errorf("End = %q", s.End)
		}
		if s.Color != 2 {
			t.Errorf("Color = %d", s.Color)
		}

		// Bitmask 127 = all days.
		if schedules[1].Day != "sunday monday tuesday wednesday thursday friday saturday" {
			t.Errorf("Day = %q, want all days", schedules[1].Day)
		}

		// Bitmask 0 = none.
		if schedules[2].Day != "none" {
			t.Errorf("Day = %q, want \"none\"", schedules[2].Day)
		}
	})
}

func TestListSchedulesOnetime(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListSchedulesOnetime(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/firewall/schedule/onetime": `[
				{
					"name": "maintenance-window",
					"start": ["22:00", "2024/03/15"],
					"end": ["06:00", "2024/03/16"],
					"color": 1
				}
			]`,
		})

		schedules, err := client.ListSchedulesOnetime(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(schedules) != 1 {
			t.Fatalf("len = %d, want 1", len(schedules))
		}

		s := schedules[0]
		if s.Name != "maintenance-window" {
			t.Errorf("Name = %q", s.Name)
		}
		if s.Type != "onetime" {
			t.Errorf("Type = %q", s.Type)
		}
		if s.Day != "" {
			t.Errorf("Day = %q, want empty for onetime", s.Day)
		}
		if s.Start != "22:00 2024/03/15" {
			t.Errorf("Start = %q", s.Start)
		}
		if s.End != "06:00 2024/03/16" {
			t.Errorf("End = %q", s.End)
		}
	})
}
