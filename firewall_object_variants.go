package fortimgr

import (
	"context"
	"fmt"
)

type apiAddress6 struct {
	Name      string `json:"name"`
	IP6       any    `json:"ip6"`
	Type      any    `json:"type"`
	Subnet    any    `json:"subnet"`
	Comment   string `json:"comment"`
	Color     int    `json:"color"`
	AssocIntf any    `json:"associated-interface"`
}

type apiVIPGroup struct {
	Name    string `json:"name"`
	Member  any    `json:"member"`
	Comment string `json:"comment"`
	Color   int    `json:"color"`
}

type apiVirtualIP6 struct {
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

type apiIPPool6 struct {
	Name    string `json:"name"`
	Type    any    `json:"type"`
	StartIP any    `json:"startip"`
	EndIP   any    `json:"endip"`
	Comment string `json:"comments"`
}

type apiIPPoolGroup struct {
	Name    string `json:"name"`
	Member  any    `json:"member"`
	Comment string `json:"comments"`
}

type apiInternetServiceCustom struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Comment string `json:"comment"`
	Entry   any    `json:"entry"`
}

type apiInternetServiceGroup struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Member  any    `json:"member"`
	Comment string `json:"comment"`
}

type apiInternetServiceName struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type apiFDSDBInternetService struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Category any    `json:"category"`
	Protocol any    `json:"protocol"`
}

type apiScheduleGroup struct {
	Name    string `json:"name"`
	Member  any    `json:"member"`
	Comment string `json:"comment"`
	Color   int    `json:"color"`
}

// ListAddresses6 retrieves IPv6 firewall address objects from an ADOM.
func (c *Client) ListAddresses6(ctx context.Context, adom string, opts ...ListOption) ([]Address6, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/address6", adom), opts, func(item apiAddress6) Address6 {
		return Address6{
			Name:      item.Name,
			Type:      mapEnum(toString(item.Type), addressTypes),
			IP6:       firstNonEmpty(toString(item.IP6), toString(item.Subnet)),
			Comment:   item.Comment,
			Color:     item.Color,
			AssocIntf: toString(item.AssocIntf),
		}
	})
}

// ListAddressGroups6 retrieves IPv6 firewall address groups from an ADOM.
func (c *Client) ListAddressGroups6(ctx context.Context, adom string, opts ...ListOption) ([]AddressGroup, error) {
	return c.listAddressGroupEndpoint(ctx, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/addrgrp6", adom), opts...)
}

// ListVirtualIPGroups retrieves IPv4 VIP groups from an ADOM.
func (c *Client) ListVirtualIPGroups(ctx context.Context, adom string, opts ...ListOption) ([]VirtualIPGroup, error) {
	return c.listVIPGroupEndpoint(ctx, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/vipgrp", adom), opts...)
}

// ListVirtualIPs6 retrieves IPv6 VIP objects from an ADOM.
func (c *Client) ListVirtualIPs6(ctx context.Context, adom string, opts ...ListOption) ([]VirtualIP6, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/vip6", adom), opts, func(item apiVirtualIP6) VirtualIP6 {
		return VirtualIP6{
			Name:        item.Name,
			ExtIP:       toString(item.ExtIP),
			MappedIP:    formatMappedIP(item.MappedIP),
			ExtIntf:     toString(item.ExtIntf),
			PortForward: mapEnum(toString(item.PortForward), enableDisable),
			Protocol:    mapEnum(toString(item.Protocol), vipProtocols),
			ExtPort:     toString(item.ExtPort),
			MappedPort:  toString(item.MappedPort),
			Comment:     item.Comment,
			Color:       item.Color,
		}
	})
}

// ListVirtualIPGroups6 retrieves IPv6 VIP groups from an ADOM.
func (c *Client) ListVirtualIPGroups6(ctx context.Context, adom string, opts ...ListOption) ([]VirtualIPGroup, error) {
	return c.listVIPGroupEndpoint(ctx, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/vipgrp6", adom), opts...)
}

// ListIPPools6 retrieves IPv6 IP pools from an ADOM.
func (c *Client) ListIPPools6(ctx context.Context, adom string, opts ...ListOption) ([]IPPool6, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/ippool6", adom), opts, func(item apiIPPool6) IPPool6 {
		return IPPool6{
			Name:    item.Name,
			Type:    mapEnum(toString(item.Type), ippoolTypes),
			StartIP: toString(item.StartIP),
			EndIP:   toString(item.EndIP),
			Comment: item.Comment,
		}
	})
}

// ListIPPoolGroups retrieves IP pool groups from an ADOM.
func (c *Client) ListIPPoolGroups(ctx context.Context, adom string, opts ...ListOption) ([]IPPoolGroup, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/ippool_grp", adom), opts, func(item apiIPPoolGroup) IPPoolGroup {
		return IPPoolGroup{Name: item.Name, Members: toStringSlice(item.Member), Comment: item.Comment}
	})
}

// ListInternetServiceCustom retrieves custom internet service objects.
func (c *Client) ListInternetServiceCustom(ctx context.Context, adom string, opts ...ListOption) ([]InternetServiceCustom, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/internet-service-custom", adom), opts, func(item apiInternetServiceCustom) InternetServiceCustom {
		return InternetServiceCustom{Name: item.Name, ID: item.ID, Comment: item.Comment, Entry: toStringSlice(item.Entry)}
	})
}

// ListInternetServiceCustomGroups retrieves custom internet service groups.
func (c *Client) ListInternetServiceCustomGroups(ctx context.Context, adom string, opts ...ListOption) ([]InternetServiceGroup, error) {
	return c.listInternetServiceGroupEndpoint(ctx, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/internet-service-custom-group", adom), opts...)
}

// ListInternetServiceGroups retrieves internet service groups.
func (c *Client) ListInternetServiceGroups(ctx context.Context, adom string, opts ...ListOption) ([]InternetServiceGroup, error) {
	return c.listInternetServiceGroupEndpoint(ctx, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/internet-service-group", adom), opts...)
}

// ListInternetServiceNames retrieves internet service names.
func (c *Client) ListInternetServiceNames(ctx context.Context, adom string, opts ...ListOption) ([]InternetServiceName, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/internet-service-name", adom), opts, func(item apiInternetServiceName) InternetServiceName {
		return InternetServiceName(item)
	})
}

// ListFDSDBInternetServices retrieves FortiGuard internet service definitions.
func (c *Client) ListFDSDBInternetServices(ctx context.Context, adom string, opts ...ListOption) ([]FDSDBInternetService, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/_fdsdb/internet-service", adom), opts, func(item apiFDSDBInternetService) FDSDBInternetService {
		return FDSDBInternetService{Name: item.Name, ID: item.ID, Category: toString(item.Category), Protocol: toString(item.Protocol)}
	})
}

// ListScheduleGroups retrieves schedule groups from an ADOM.
func (c *Client) ListScheduleGroups(ctx context.Context, adom string, opts ...ListOption) ([]ScheduleGroup, error) {
	return listADOMEndpoint(ctx, c, adom, fmt.Sprintf("/pm/config/adom/%s/obj/firewall/schedule/group", adom), opts, func(item apiScheduleGroup) ScheduleGroup {
		return ScheduleGroup{Name: item.Name, Members: toStringSlice(item.Member), Comment: item.Comment, Color: item.Color}
	})
}

func (c *Client) listAddressGroupEndpoint(ctx context.Context, adom, apiURL string, opts ...ListOption) ([]AddressGroup, error) {
	return listADOMEndpoint(ctx, c, adom, apiURL, opts, func(item apiAddressGroup) AddressGroup {
		return AddressGroup{Name: item.Name, Members: toStringSlice(item.Member), Comment: item.Comment, Color: item.Color}
	})
}

func (c *Client) listVIPGroupEndpoint(ctx context.Context, adom, apiURL string, opts ...ListOption) ([]VirtualIPGroup, error) {
	return listADOMEndpoint(ctx, c, adom, apiURL, opts, func(item apiVIPGroup) VirtualIPGroup {
		return VirtualIPGroup{Name: item.Name, Members: toStringSlice(item.Member), Comment: item.Comment, Color: item.Color}
	})
}

func (c *Client) listInternetServiceGroupEndpoint(ctx context.Context, adom, apiURL string, opts ...ListOption) ([]InternetServiceGroup, error) {
	return listADOMEndpoint(ctx, c, adom, apiURL, opts, func(item apiInternetServiceGroup) InternetServiceGroup {
		return InternetServiceGroup{Name: item.Name, ID: item.ID, Members: toStringSlice(item.Member), Comment: item.Comment}
	})
}

func listADOMEndpoint[T, R any](ctx context.Context, c *Client, adom, apiURL string, opts []ListOption, convert func(T) R) ([]R, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}
	items, err := getPaged[T](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}
	out := make([]R, len(items))
	for i, item := range items {
		out[i] = convert(item)
	}
	return out, nil
}
