package fortimgr

import (
	"context"
	"fmt"
)

type apiService struct {
	Name         string `json:"name"`
	Protocol     any    `json:"protocol"`
	TCPPortRange any    `json:"tcp-portrange"`
	UDPPortRange any    `json:"udp-portrange"`
	Comment      string `json:"comment"`
}

type apiServiceGroup struct {
	Name    string `json:"name"`
	Member  any    `json:"member"`
	Comment string `json:"comment"`
}

// ListServices retrieves firewall service objects from an ADOM.
func (c *Client) ListServices(ctx context.Context, adom string) ([]Service, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/service/custom", adom)
	items, err := get[apiService](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	services := make([]Service, len(items))
	for i, s := range items {
		services[i] = Service{
			Name:     s.Name,
			Protocol: mapEnum(toString(s.Protocol), serviceProtocols),
			TCPRange: toString(s.TCPPortRange),
			UDPRange: toString(s.UDPPortRange),
			Comment:  s.Comment,
		}
	}

	return services, nil
}

// ListServiceGroups retrieves firewall service groups from an ADOM.
func (c *Client) ListServiceGroups(ctx context.Context, adom string) ([]ServiceGroup, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/service/group", adom)
	items, err := get[apiServiceGroup](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	groups := make([]ServiceGroup, len(items))
	for i, g := range items {
		groups[i] = ServiceGroup{
			Name:    g.Name,
			Members: toStringSlice(g.Member),
			Comment: g.Comment,
		}
	}

	return groups, nil
}
