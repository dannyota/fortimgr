package fortimgr

import (
	"context"
	"fmt"
)

type apiVirtualIP struct {
	Name        string `json:"name"`
	ExtIP       any    `json:"extip"`
	MappedIP    any    `json:"mappedip"`
	ExtIntf     any    `json:"extintf"`
	PortForward any    `json:"portforward"`
	Protocol    any    `json:"protocol"`
	ExtPort     any    `json:"extport"`
	MappedPort  any    `json:"mappedport"`
	Comment     string `json:"comment"`
	Color       int    `json:"color"`
}

// ListVirtualIPs retrieves Virtual IPs from an ADOM.
func (c *Client) ListVirtualIPs(ctx context.Context, adom string) ([]VirtualIP, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/vip", adom)
	items, err := get[apiVirtualIP](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	virtualIPs := make([]VirtualIP, len(items))
	for i, v := range items {
		virtualIPs[i] = VirtualIP{
			Name:        v.Name,
			ExtIP:       toString(v.ExtIP),
			MappedIP:    formatMappedIP(v.MappedIP),
			ExtIntf:     toString(v.ExtIntf),
			PortForward: mapEnum(toString(v.PortForward), enableDisable),
			Protocol:    mapEnum(toString(v.Protocol), vipProtocols),
			ExtPort:     toString(v.ExtPort),
			MappedPort:  toString(v.MappedPort),
			Comment:     v.Comment,
			Color:       v.Color,
		}
	}

	return virtualIPs, nil
}
