package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
)

type apiDeviceFirmware struct {
	Name           string `json:"name"`
	DevSN          string `json:"devsn"`
	Platform       string `json:"platform_str"`
	CurrVer        string `json:"curr_ver"`
	CurrBuild      int    `json:"curr_build"`
	UpdVer         string `json:"upd_ver"`
	CanUpgrade     int    `json:"can_upgrade"`
	Connection     int    `json:"connection"`
	IsLicenseValid int    `json:"is_license_valid"`
	IsGroup        int    `json:"isGroup"`
}

// ListDeviceFirmware retrieves firmware status for all managed devices via flatui_proxy.
func (c *Client) ListDeviceFirmware(ctx context.Context) ([]DeviceFirmware, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.proxy(ctx, "/gui/adom/dvm/firmware/management", "loadFirmwareDataGroupByVersion")
	if err != nil {
		return nil, err
	}

	var items []apiDeviceFirmware
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal firmware: %w", err)
	}

	var result []DeviceFirmware
	for _, f := range items {
		if f.IsGroup == 1 {
			continue // skip group headers
		}
		result = append(result, DeviceFirmware{
			Name:           f.Name,
			SerialNumber:   f.DevSN,
			Platform:       f.Platform,
			CurrentVersion: f.CurrVer,
			CurrentBuild:   f.CurrBuild,
			UpgradeVersion: f.UpdVer,
			CanUpgrade:     f.CanUpgrade == 1,
			Connected:      f.Connection == 1,
			LicenseValid:   f.IsLicenseValid == 1,
		})
	}

	return result, nil
}
