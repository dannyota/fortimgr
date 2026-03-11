package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type apiSystemStatus struct {
	Hostname    string `json:"hostname"`
	SN          string `json:"sn"`
	FMGVersion  string `json:"fmgversion"`
	BuildNumber int    `json:"build_number"`
	HAMode      int    `json:"ha_mode"`
	PlatformID  string `json:"platform-id"`
}

// SystemStatus retrieves the FortiManager system status via flatui_proxy.
func (c *Client) SystemStatus(ctx context.Context) (*SystemStatus, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.proxy(ctx, "/gui/sys/config", "get")
	if err != nil {
		return nil, err
	}

	var item apiSystemStatus
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal system status: %w", err)
	}

	return &SystemStatus{
		Hostname:     item.Hostname,
		SerialNumber: item.SN,
		Version:      item.FMGVersion,
		Platform:     item.PlatformID,
		Build:        item.BuildNumber,
		HAMode:       mapEnum(strconv.Itoa(item.HAMode), deviceHAModes),
	}, nil
}
