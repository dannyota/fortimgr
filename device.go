package fortimgr

import (
	"context"
	"fmt"
)

type apiDevice struct {
	Name        string       `json:"name"`
	OID         int          `json:"oid"`
	SN          string       `json:"sn"`
	Platform    string       `json:"platform_str"`
	OSVer       int          `json:"os_ver"`
	MrNum       int          `json:"mr"`
	PatchNum    int          `json:"patch"`
	BuildNum    int          `json:"build"`
	HAMode      any          `json:"ha_mode"`
	HACluster   int          `json:"ha_cluster"`
	ConnState   any          `json:"conn_status"`
	IP          string       `json:"ip"`
	Hostname    string       `json:"hostname"`
	ConfStatus  any          `json:"conf_status"`
	DevStatus   any          `json:"dev_status"`
	LastChecked any          `json:"last_checked"`
	LastResync  any          `json:"last_resync"`
	HASlave     []apiHASlave `json:"ha_slave"`

	// v1.1.0 — license / subscription fields. These map to the 10 flat
	// License* fields on Device. adm_pass, private_key, and psk are
	// deliberately NOT declared here so they cannot be unmarshaled even
	// if FortiManager includes them in the response.
	VMLicExpire       any    `json:"vm_lic_expire"`
	VMLicOverdueSince any    `json:"vm_lic_overdue_since"`
	FoslicCPU         int    `json:"foslic_cpu"`
	FoslicRAM         int    `json:"foslic_ram"`
	FoslicUTM         int    `json:"foslic_utm"`
	FoslicType        int    `json:"foslic_type"`
	FoslicDRSite      int    `json:"foslic_dr_site"`
	FoslicInstTime    any    `json:"foslic_inst_time"`
	FoslicLastSync    any    `json:"foslic_last_sync"`
	LicFlags          int    `json:"lic_flags"`
	LicRegion         string `json:"lic_region"`
}

// deviceFields is the allowlist of fields ListDevices requests from
// /dvmdb/adom/{adom}/device. Explicitly excludes adm_pass, private_key,
// and psk so encrypted device credentials never transit the wire from
// FortiManager to the SDK process. Verified against a live FortiManager:
// requesting the allowlist returns 30 fields vs 102 without it, and
// none of the 4 known credential fields appear in the filtered response.
var deviceFields = []string{
	// identity
	"name", "hostname", "oid", "sn", "platform_str", "ip",
	// firmware
	"os_ver", "mr", "patch", "build",
	// operational status
	"conn_status", "conf_status", "dev_status", "last_checked", "last_resync",
	// HA
	"ha_mode", "ha_cluster", "ha_slave",
	// license / subscription
	"vm_lic_expire", "vm_lic_overdue_since",
	"foslic_cpu", "foslic_ram", "foslic_utm", "foslic_type", "foslic_dr_site",
	"foslic_inst_time", "foslic_last_sync",
	"lic_flags", "lic_region",
}

// apiHASlave is the per-member entry in the dvmdb device's ha_slave array.
// Each entry describes one FortiGate in the HA cluster. role=1 is the
// primary/master; role=0 is the secondary/slave.
type apiHASlave struct {
	Name       string `json:"name"`
	Role       any    `json:"role"`
	SN         string `json:"sn"`
	Status     any    `json:"status"`
	ConfStatus any    `json:"conf_status"`
}

// ListDevices retrieves all FortiGate devices from an ADOM.
// An ADOM (Administrative Domain) is FortiManager's multi-tenancy scope.
// Use "root" for the default ADOM in single-tenant deployments.
//
// Pagination: fetches every page (1000 rows default) via getPaged,
// merging the deviceFields allowlist into every per-page request so
// encrypted credentials are never fetched from the wire. Override page
// size or observe progress via WithPageSize / WithPageCallback.
func (c *Client) ListDevices(ctx context.Context, adom string, opts ...ListOption) ([]Device, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/dvmdb/adom/%s/device", adom)
	items, err := getPaged[apiDevice](ctx, c, apiURL, map[string]any{
		"fields": deviceFields,
	}, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	devices := make([]Device, len(items))
	for i, d := range items {
		devices[i] = Device{
			Name:         d.Name,
			DeviceID:     fmt.Sprintf("%d", d.OID),
			SerialNumber: d.SN,
			Platform:     d.Platform,
			Firmware:     fmt.Sprintf("%d.%d.%d-b%d", d.OSVer, d.MrNum, d.PatchNum, d.BuildNum),
			HAMode:       mapEnum(toString(d.HAMode), deviceHAModes),
			HATopology:   mapEnum(toString(d.HAMode), deviceHATopologies),
			HAClusterID:  fmt.Sprintf("%d", d.HACluster),
			Status:       mapEnum(toString(d.ConnState), deviceConnStates),
			IPAddress:    d.IP,
			ADOM:         adom,

			Hostname:    d.Hostname,
			ConfStatus:  mapEnum(toString(d.ConfStatus), confStatuses),
			DevStatus:   mapEnum(toString(d.DevStatus), devStatuses),
			LastChecked: unixToTime(d.LastChecked),
			LastResync:  unixToTime(d.LastResync),
			HARole:      haRoleFromSlaves(d.Name, d.HASlave),
			HAMembers:   haMembers(d.HASlave),

			LicenseExpire:       unixToTime(d.VMLicExpire),
			LicenseOverdueSince: unixToTime(d.VMLicOverdueSince),
			LicenseMaxCPU:       d.FoslicCPU,
			LicenseMaxRAM:       d.FoslicRAM,
			LicenseUTMEnabled:   d.FoslicUTM != 0,
			LicenseType:         d.FoslicType,
			LicenseInstalledAt:  unixToTime(d.FoslicInstTime),
			LicenseLastSync:     unixToTime(d.FoslicLastSync),
			LicenseRegion:       d.LicRegion,
			LicenseFlags:        d.LicFlags,
		}
	}

	return devices, nil
}

// haMembers converts the raw ha_slave[] array into a slice of HAMember.
// Returns nil for standalone devices (empty ha_slave).
func haMembers(slaves []apiHASlave) []HAMember {
	if len(slaves) == 0 {
		return nil
	}
	members := make([]HAMember, len(slaves))
	for i, s := range slaves {
		members[i] = HAMember{
			Name:         s.Name,
			SerialNumber: s.SN,
			Role:         mapEnum(toString(s.Role), haRoles),
			Status:       mapEnum(toString(s.Status), deviceConnStates),
			ConfStatus:   mapEnum(toString(s.ConfStatus), confStatuses),
		}
	}
	return members
}

// haRoleFromSlaves returns the role of the named device within its HA
// cluster, derived from the ha_slave[] array. Empty string when the device
// is standalone or no matching slave entry exists.
func haRoleFromSlaves(name string, slaves []apiHASlave) string {
	if len(slaves) == 0 {
		return ""
	}
	for _, s := range slaves {
		if s.Name == name {
			return mapEnum(toString(s.Role), haRoles)
		}
	}
	return ""
}
