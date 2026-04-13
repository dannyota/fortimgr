package fortimgr

import (
	"context"
	"fmt"
)

type apiScheduleRecurring struct {
	Name  string `json:"name"`
	Day   any    `json:"day"`
	Start any    `json:"start"`
	End   any    `json:"end"`
	Color int    `json:"color"`
}

type apiScheduleOnetime struct {
	Name  string `json:"name"`
	Start any    `json:"start"`
	End   any    `json:"end"`
	Color int    `json:"color"`
}

// ListSchedulesRecurring retrieves recurring schedules from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListSchedulesRecurring(ctx context.Context, adom string, opts ...ListOption) ([]Schedule, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/schedule/recurring", adom)
	items, err := getPaged[apiScheduleRecurring](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	schedules := make([]Schedule, len(items))
	for i, s := range items {
		schedules[i] = Schedule{
			Name:  s.Name,
			Type:  "recurring",
			Day:   mapScheduleDay(toString(s.Day)),
			Start: formatScheduleTime(s.Start),
			End:   formatScheduleTime(s.End),
			Color: s.Color,
		}
	}

	return schedules, nil
}

// ListSchedulesOnetime retrieves one-time schedules from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListSchedulesOnetime(ctx context.Context, adom string, opts ...ListOption) ([]Schedule, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/schedule/onetime", adom)
	items, err := getPaged[apiScheduleOnetime](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	schedules := make([]Schedule, len(items))
	for i, s := range items {
		schedules[i] = Schedule{
			Name:  s.Name,
			Type:  "onetime",
			Start: formatScheduleTime(s.Start),
			End:   formatScheduleTime(s.End),
			Color: s.Color,
		}
	}

	return schedules, nil
}
