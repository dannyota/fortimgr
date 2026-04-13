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
func (c *Client) ListDevices(ctx context.Context, adom string) ([]Device, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/dvmdb/adom/%s/device", adom)
	items, err := get[apiDevice](ctx, c, apiURL)
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
