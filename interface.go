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

// ListInterfaces retrieves network interfaces from a device VDOM.
func (c *Client) ListInterfaces(ctx context.Context, device, vdom string) ([]Interface, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}

	apiURL := fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/interface", device, vdom)
	items, err := get[apiInterface](ctx, c, apiURL)
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
