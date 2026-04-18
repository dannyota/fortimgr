package fortimgr

import (
	"context"
	"fmt"
)

type apiStaticRoute struct {
	SeqNum   int    `json:"seq-num"`
	Dst      any    `json:"dst"`
	Gateway  string `json:"gateway"`
	Device   any    `json:"device"`
	Distance int    `json:"distance"`
	Priority int    `json:"priority"`
	Status   any    `json:"status"`
	Comment  string `json:"comment"`
}

type apiStaticRoute6 struct {
	SeqNum   int    `json:"seq-num"`
	Dst      any    `json:"dst"`
	Gateway  string `json:"gateway"`
	Device   any    `json:"device"`
	Distance int    `json:"distance"`
	Priority int    `json:"priority"`
	Status   any    `json:"status"`
	Comment  string `json:"comment"`
}

// ListStaticRoutes retrieves static routes from a device VDOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListStaticRoutes(ctx context.Context, device, vdom string, opts ...ListOption) ([]StaticRoute, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}

	apiURL := fmt.Sprintf("/pm/config/device/%s/vdom/%s/router/static", device, vdom)
	items, err := getPaged[apiStaticRoute](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	routes := make([]StaticRoute, len(items))
	for i, r := range items {
		routes[i] = StaticRoute{
			SeqNum:   r.SeqNum,
			Dst:      formatSubnet(r.Dst),
			Gateway:  r.Gateway,
			Device:   toString(r.Device),
			Distance: r.Distance,
			Priority: r.Priority,
			Status:   mapEnum(toString(r.Status), enableDisable),
			Comment:  r.Comment,
		}
	}

	return routes, nil
}

// ListStaticRoutes6 retrieves IPv6 static routes from a device VDOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListStaticRoutes6(ctx context.Context, device, vdom string, opts ...ListOption) ([]StaticRoute6, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}

	apiURL := fmt.Sprintf("/pm/config/device/%s/vdom/%s/router/static6", device, vdom)
	items, err := getPaged[apiStaticRoute6](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	routes := make([]StaticRoute6, len(items))
	for i, r := range items {
		routes[i] = StaticRoute6{
			SeqNum:   r.SeqNum,
			Dst:      toString(r.Dst),
			Gateway:  r.Gateway,
			Device:   toString(r.Device),
			Distance: r.Distance,
			Priority: r.Priority,
			Status:   mapEnum(toString(r.Status), enableDisable),
			Comment:  r.Comment,
		}
	}

	return routes, nil
}
