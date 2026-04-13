package fortimgr

import (
	"context"
	"fmt"
	"strings"
)

type apiInterface struct {
	Name        string `json:"name"`
	IP          any    `json:"ip"`
	Type        any    `json:"type"`
	Status      any    `json:"status"`
	Role        any    `json:"role"`
	Mode        any    `json:"mode"`
	AllowAccess any    `json:"allowaccess"`
	VDOM        any    `json:"vdom"`
	Zone        any    `json:"zone"`
	VlanID      int    `json:"vlanid"`
	MTU         int    `json:"mtu"`
	Speed       any    `json:"speed"`
	Alias       string `json:"alias"`
	Description string `json:"description"`
}

// ListInterfaces retrieves network interfaces for a device.
//
// Pass an empty string (or "global") for vdom to use the device-wide global
// path (/pm/config/device/<dev>/global/system/interface). This returns every
// interface across all VDOMs in a single call, with each interface carrying
// its own "vdom" field — callers can derive the VDOM set from the result.
// This is the only path available to restricted admins that cannot enumerate
// /dvmdb/device/<dev>/vdom.
//
// Pass a specific vdom name to scope the query to that VDOM.
//
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListInterfaces(ctx context.Context, device, vdom string, opts ...ListOption) ([]Interface, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) {
		return nil, fmt.Errorf("%w: device=%q", ErrInvalidName, device)
	}

	var apiURL string
	if vdom == "" || vdom == "global" {
		apiURL = fmt.Sprintf("/pm/config/device/%s/global/system/interface", device)
	} else {
		if !validName(vdom) {
			return nil, fmt.Errorf("%w: vdom=%q", ErrInvalidName, vdom)
		}
		apiURL = fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/interface", device, vdom)
	}
	items, err := getPaged[apiInterface](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	interfaces := make([]Interface, len(items))
	for i, iface := range items {
		interfaces[i] = Interface{
			Name:        iface.Name,
			IP:          formatSubnet(iface.IP),
			Type:        mapEnum(toString(iface.Type), interfaceTypes),
			Status:      mapEnum(toString(iface.Status), interfaceStatuses),
			Role:        mapEnum(toString(iface.Role), interfaceRoles),
			Mode:        mapEnum(toString(iface.Mode), interfaceModes),
			AllowAccess: strings.Join(toStringSlice(iface.AllowAccess), " "),
			VDOM:        toString(iface.VDOM),
			Zone:        toString(iface.Zone),
			VlanID:      iface.VlanID,
			MTU:         iface.MTU,
			Speed:       toString(iface.Speed),
			Alias:       iface.Alias,
			Description: iface.Description,
		}
	}

	return interfaces, nil
}
