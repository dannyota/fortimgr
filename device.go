package fortimgr

import (
	"context"
	"fmt"
)

type apiDevice struct {
	Name      string `json:"name"`
	OID       int    `json:"oid"`
	SN        string `json:"sn"`
	Platform  string `json:"platform_str"`
	OSVer     int    `json:"os_ver"`
	MrNum     int    `json:"mr"`
	PatchNum  int    `json:"patch"`
	BuildNum  int    `json:"build"`
	HAMode    any    `json:"ha_mode"`
	HACluster int    `json:"ha_cluster"`
	ConnState any    `json:"conn_status"`
	IP        string `json:"ip"`
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
		}
	}

	return devices, nil
}
