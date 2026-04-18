package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
)

type apiDeviceDNS struct {
	OID                    int      `json:"oid"`
	Primary                string   `json:"primary"`
	Secondary              string   `json:"secondary"`
	AltPrimary             string   `json:"alt-primary"`
	AltSecondary           string   `json:"alt-secondary"`
	IPv6Primary            string   `json:"ip6-primary"`
	IPv6Secondary          string   `json:"ip6-secondary"`
	Protocol               any      `json:"protocol"`
	ServerSelectMethod     any      `json:"server-select-method"`
	ServerHostname         []string `json:"server-hostname"`
	Domain                 []string `json:"domain"`
	InterfaceSelectMethod  any      `json:"interface-select-method"`
	Interfaces             []string `json:"interface"`
	SourceIP               string   `json:"source-ip"`
	Retry                  int      `json:"retry"`
	Timeout                int      `json:"timeout"`
	DNSCacheLimit          int      `json:"dns-cache-limit"`
	DNSCacheTTL            int      `json:"dns-cache-ttl"`
	FQDNCacheTTL           int      `json:"fqdn-cache-ttl"`
	FQDNMinRefresh         int      `json:"fqdn-min-refresh"`
	FQDNMaxRefresh         int      `json:"fqdn-max-refresh"`
	CacheNotFoundResponses any      `json:"cache-notfound-responses"`
	Log                    any      `json:"log"`
	SSLCertificate         []string `json:"ssl-certificate"`
}

type apiDeviceDDNS struct {
	Name             string `json:"name"`
	DDNSDomain       string `json:"ddns-domain"`
	MonitorInterface any    `json:"monitor-interface"`
	Server           any    `json:"server"`
	Status           any    `json:"status"`
	Username         string `json:"username"`
}

type apiDeviceIPAM struct {
	OID                         int `json:"oid"`
	Status                      any `json:"status"`
	ServerType                  any `json:"server-type"`
	AutomaticConflictResolution any `json:"automatic-conflict-resolution"`
	ManageLANAddresses          any `json:"manage-lan-addresses"`
	ManageLANExtensionAddresses any `json:"manage-lan-extension-addresses"`
	ManageSSIDAddresses         any `json:"manage-ssid-addresses"`
	RequireSubnetSizeMatch      any `json:"require-subnet-size-match"`
}

type apiSDWANSettings struct {
	OID                    int                   `json:"oid"`
	Status                 any                   `json:"status"`
	LoadBalanceMode        any                   `json:"load-balance-mode"`
	FailDetect             any                   `json:"fail-detect"`
	AppPerfLogPeriod       int                   `json:"app-perf-log-period"`
	DuplicationMaxNum      int                   `json:"duplication-max-num"`
	NeighborHoldBootTime   int                   `json:"neighbor-hold-boot-time"`
	NeighborHoldDown       any                   `json:"neighbor-hold-down"`
	NeighborHoldDownTime   int                   `json:"neighbor-hold-down-time"`
	SpeedtestBypassRouting any                   `json:"speedtest-bypass-routing"`
	FailAlertInterfaces    []string              `json:"fail-alert-interfaces"`
	HealthChecks           []apiSDWANHealthCheck `json:"health-check"`
	Zones                  []apiSDWANZone        `json:"zone"`
}

type apiSDWANHealthCheck struct {
	Name                   string        `json:"name"`
	OID                    int           `json:"oid"`
	Protocol               any           `json:"protocol"`
	Server                 []string      `json:"server"`
	Members                []int         `json:"members"`
	Interval               int           `json:"interval"`
	Failtime               int           `json:"failtime"`
	Recoverytime           int           `json:"recoverytime"`
	Source                 string        `json:"source"`
	Source6                string        `json:"source6"`
	SystemDNS              any           `json:"system-dns"`
	UpdateStaticRoute      any           `json:"update-static-route"`
	UpdateCascadeInterface any           `json:"update-cascade-interface"`
	SLA                    []apiSDWANSLA `json:"sla"`
}

type apiSDWANSLA struct {
	ID                  int `json:"id"`
	OID                 int `json:"oid"`
	LatencyThreshold    int `json:"latency-threshold"`
	JitterThreshold     int `json:"jitter-threshold"`
	PacketLossThreshold int `json:"packetloss-threshold"`
	LinkCostFactor      any `json:"link-cost-factor"`
}

type apiSDWANZone struct {
	Name                  string `json:"name"`
	OID                   int    `json:"oid"`
	MinimumSLAMeetMembers int    `json:"minimum-sla-meet-members"`
	ServiceSLATieBreak    any    `json:"service-sla-tie-break"`
	ADVPNSelect           any    `json:"advpn-select"`
}

type apiSDWANMember struct {
	SeqNum    int    `json:"seq-num"`
	Interface string `json:"interface"`
	Gateway   string `json:"gateway"`
	Gateway6  string `json:"gateway6"`
	Zone      string `json:"zone"`
	Status    any    `json:"status"`
	Cost      int    `json:"cost"`
	Priority  int    `json:"priority"`
	Comment   string `json:"comment"`
}

type apiSDWANService struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Mode            any    `json:"mode"`
	Status          any    `json:"status"`
	AddressMode     any    `json:"addr-mode"`
	InputDevice     any    `json:"input-device"`
	Src             any    `json:"src"`
	Dst             any    `json:"dst"`
	InternetService any    `json:"internet-service"`
	PriorityMembers any    `json:"priority-members"`
}

type apiSDWANDuplication struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	ServiceID         int    `json:"service-id"`
	PacketDuplication any    `json:"packet-duplication"`
	Fields            any    `json:"fields"`
}

var sdwanSettingsFields = []string{
	"oid", "status", "load-balance-mode", "fail-detect",
	"app-perf-log-period", "duplication-max-num",
	"neighbor-hold-boot-time", "neighbor-hold-down", "neighbor-hold-down-time",
	"speedtest-bypass-routing", "fail-alert-interfaces", "health-check", "zone",
}

// DeviceDNS retrieves global DNS settings from a managed device.
func (c *Client) DeviceDNS(ctx context.Context, device string) (*DeviceDNS, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) {
		return nil, fmt.Errorf("%w: device=%q", ErrInvalidName, device)
	}

	var raw apiDeviceDNS
	if err := c.forwardObject(ctx, fmt.Sprintf("/pm/config/device/%s/global/system/dns", device), nil, &raw); err != nil {
		return nil, err
	}
	out := DeviceDNS{
		OID:                    raw.OID,
		Primary:                raw.Primary,
		Secondary:              raw.Secondary,
		AltPrimary:             raw.AltPrimary,
		AltSecondary:           raw.AltSecondary,
		IPv6Primary:            raw.IPv6Primary,
		IPv6Secondary:          raw.IPv6Secondary,
		Protocol:               toString(raw.Protocol),
		ServerSelectMethod:     toString(raw.ServerSelectMethod),
		ServerHostname:         raw.ServerHostname,
		Domain:                 raw.Domain,
		InterfaceSelectMethod:  toString(raw.InterfaceSelectMethod),
		Interfaces:             raw.Interfaces,
		SourceIP:               raw.SourceIP,
		Retry:                  raw.Retry,
		Timeout:                raw.Timeout,
		DNSCacheLimit:          raw.DNSCacheLimit,
		DNSCacheTTL:            raw.DNSCacheTTL,
		FQDNCacheTTL:           raw.FQDNCacheTTL,
		FQDNMinRefresh:         raw.FQDNMinRefresh,
		FQDNMaxRefresh:         raw.FQDNMaxRefresh,
		CacheNotFoundResponses: mapEnum(toString(raw.CacheNotFoundResponses), enableDisable),
		Log:                    mapEnum(toString(raw.Log), enableDisable),
		SSLCertificate:         raw.SSLCertificate,
	}
	return &out, nil
}

// ListDeviceDDNS retrieves global DDNS settings from a managed device.
func (c *Client) ListDeviceDDNS(ctx context.Context, device string, opts ...ListOption) ([]DeviceDDNS, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) {
		return nil, fmt.Errorf("%w: device=%q", ErrInvalidName, device)
	}
	items, err := getPaged[apiDeviceDDNS](ctx, c, fmt.Sprintf("/pm/config/device/%s/global/system/ddns", device), nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}
	out := make([]DeviceDDNS, len(items))
	for i, item := range items {
		out[i] = DeviceDDNS{
			Name:             item.Name,
			DDNSDomain:       item.DDNSDomain,
			MonitorInterface: toString(item.MonitorInterface),
			Server:           toString(item.Server),
			Status:           mapEnum(toString(item.Status), enableDisable),
			Username:         item.Username,
		}
	}
	return out, nil
}

// DeviceIPAM retrieves IPAM settings from a device VDOM.
func (c *Client) DeviceIPAM(ctx context.Context, device, vdom string) (*DeviceIPAM, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}
	var raw apiDeviceIPAM
	if err := c.forwardObject(ctx, fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/ipam", device, vdom), nil, &raw); err != nil {
		return nil, err
	}
	out := DeviceIPAM{
		OID:                         raw.OID,
		Status:                      mapEnum(toString(raw.Status), enableDisable),
		ServerType:                  toString(raw.ServerType),
		AutomaticConflictResolution: mapEnum(toString(raw.AutomaticConflictResolution), enableDisable),
		ManageLANAddresses:          mapEnum(toString(raw.ManageLANAddresses), enableDisable),
		ManageLANExtensionAddresses: mapEnum(toString(raw.ManageLANExtensionAddresses), enableDisable),
		ManageSSIDAddresses:         mapEnum(toString(raw.ManageSSIDAddresses), enableDisable),
		RequireSubnetSizeMatch:      mapEnum(toString(raw.RequireSubnetSizeMatch), enableDisable),
	}
	return &out, nil
}

// SDWANSettings retrieves SD-WAN settings from a device VDOM. Sensitive fields
// such as health-check passwords are deliberately omitted from the request
// allowlist and public model.
func (c *Client) SDWANSettings(ctx context.Context, device, vdom string) (*SDWANSettings, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}
	var raw apiSDWANSettings
	if err := c.forwardObject(ctx, fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan", device, vdom), map[string]any{"fields": sdwanSettingsFields}, &raw); err != nil {
		return nil, err
	}
	out := convertSDWANSettings(raw)
	return &out, nil
}

// ListSDWANMembers retrieves SD-WAN members from a device VDOM.
func (c *Client) ListSDWANMembers(ctx context.Context, device, vdom string, opts ...ListOption) ([]SDWANMember, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}
	items, err := getPaged[apiSDWANMember](ctx, c, fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/members", device, vdom), nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}
	out := make([]SDWANMember, len(items))
	for i, item := range items {
		out[i] = SDWANMember{
			SeqNum:    item.SeqNum,
			Interface: item.Interface,
			Gateway:   item.Gateway,
			Gateway6:  item.Gateway6,
			Zone:      item.Zone,
			Status:    mapEnum(toString(item.Status), enableDisable),
			Cost:      item.Cost,
			Priority:  item.Priority,
			Comment:   item.Comment,
		}
	}
	return out, nil
}

// ListSDWANServices retrieves SD-WAN service rules from a device VDOM.
func (c *Client) ListSDWANServices(ctx context.Context, device, vdom string, opts ...ListOption) ([]SDWANService, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}
	items, err := getPaged[apiSDWANService](ctx, c, fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/service", device, vdom), nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}
	out := make([]SDWANService, len(items))
	for i, item := range items {
		out[i] = SDWANService{
			ID:              item.ID,
			Name:            item.Name,
			Mode:            toString(item.Mode),
			Status:          mapEnum(toString(item.Status), enableDisable),
			AddressMode:     toString(item.AddressMode),
			InputDevice:     toStringSlice(item.InputDevice),
			Source:          toStringSlice(item.Src),
			Destination:     toStringSlice(item.Dst),
			InternetService: mapEnum(toString(item.InternetService), enableDisable),
			PriorityMembers: toStringSlice(item.PriorityMembers),
		}
	}
	return out, nil
}

// ListSDWANDuplication retrieves SD-WAN duplication rules from a device VDOM.
func (c *Client) ListSDWANDuplication(ctx context.Context, device, vdom string, opts ...ListOption) ([]SDWANDuplication, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) || !validName(vdom) {
		return nil, fmt.Errorf("%w: device=%q vdom=%q", ErrInvalidName, device, vdom)
	}
	items, err := getPaged[apiSDWANDuplication](ctx, c, fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/duplication", device, vdom), nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}
	out := make([]SDWANDuplication, len(items))
	for i, item := range items {
		out[i] = SDWANDuplication{
			ID:                item.ID,
			Name:              item.Name,
			ServiceID:         item.ServiceID,
			PacketDuplication: toString(item.PacketDuplication),
			Fields:            toStringSlice(item.Fields),
		}
	}
	return out, nil
}

func (c *Client) forwardObject(ctx context.Context, apiURL string, extra map[string]any, out any) error {
	var (
		data json.RawMessage
		err  error
	)
	if extra == nil {
		data, err = c.forward(ctx, apiURL)
	} else {
		data, err = c.forwardExtra(ctx, apiURL, extra)
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("fortimgr: unmarshal response: %w", err)
	}
	return nil
}

func convertSDWANSettings(raw apiSDWANSettings) SDWANSettings {
	out := SDWANSettings{
		OID:                    raw.OID,
		Status:                 mapEnum(toString(raw.Status), enableDisable),
		LoadBalanceMode:        toString(raw.LoadBalanceMode),
		FailDetect:             mapEnum(toString(raw.FailDetect), enableDisable),
		AppPerfLogPeriod:       raw.AppPerfLogPeriod,
		DuplicationMaxNum:      raw.DuplicationMaxNum,
		NeighborHoldBootTime:   raw.NeighborHoldBootTime,
		NeighborHoldDown:       mapEnum(toString(raw.NeighborHoldDown), enableDisable),
		NeighborHoldDownTime:   raw.NeighborHoldDownTime,
		SpeedtestBypassRouting: mapEnum(toString(raw.SpeedtestBypassRouting), enableDisable),
		FailAlertInterfaces:    raw.FailAlertInterfaces,
		Zones:                  make([]SDWANZone, len(raw.Zones)),
		HealthChecks:           make([]SDWANHealthCheck, len(raw.HealthChecks)),
	}
	for i, zone := range raw.Zones {
		out.Zones[i] = SDWANZone{
			Name:                  zone.Name,
			OID:                   zone.OID,
			MinimumSLAMeetMembers: zone.MinimumSLAMeetMembers,
			ServiceSLATieBreak:    toString(zone.ServiceSLATieBreak),
			ADVPNSelect:           mapEnum(toString(zone.ADVPNSelect), enableDisable),
		}
	}
	for i, health := range raw.HealthChecks {
		out.HealthChecks[i] = convertSDWANHealthCheck(health)
	}
	return out
}

func convertSDWANHealthCheck(raw apiSDWANHealthCheck) SDWANHealthCheck {
	out := SDWANHealthCheck{
		Name:                   raw.Name,
		OID:                    raw.OID,
		Protocol:               toString(raw.Protocol),
		Server:                 raw.Server,
		Members:                raw.Members,
		Interval:               raw.Interval,
		Failtime:               raw.Failtime,
		Recoverytime:           raw.Recoverytime,
		Source:                 raw.Source,
		Source6:                raw.Source6,
		SystemDNS:              mapEnum(toString(raw.SystemDNS), enableDisable),
		UpdateStaticRoute:      mapEnum(toString(raw.UpdateStaticRoute), enableDisable),
		UpdateCascadeInterface: mapEnum(toString(raw.UpdateCascadeInterface), enableDisable),
		SLA:                    make([]SDWANSLA, len(raw.SLA)),
	}
	for i, sla := range raw.SLA {
		out.SLA[i] = SDWANSLA{
			ID:                  sla.ID,
			OID:                 sla.OID,
			LatencyThreshold:    sla.LatencyThreshold,
			JitterThreshold:     sla.JitterThreshold,
			PacketLossThreshold: sla.PacketLossThreshold,
			LinkCostFactor:      toString(sla.LinkCostFactor),
		}
	}
	return out
}
