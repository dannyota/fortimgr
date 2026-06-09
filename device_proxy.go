package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
)

// DeviceStatus holds system status from a managed FortiGate, including uptime.
type DeviceStatus struct {
	Hostname     string `json:"hostname"`
	Serial       string `json:"serial"`
	Version      string `json:"version"`
	Build        int    `json:"build"`
	Uptime       int64  `json:"uptime"`
	ModelName    string `json:"model_name"`
	Model        string `json:"model"`
	LogDiskUsage int    `json:"log_disk_usage"`
}

// DeviceResourceUsage holds CPU and memory usage from a managed FortiGate.
type DeviceResourceUsage struct {
	CPU    []DeviceResourceSample `json:"cpu"`
	Memory []DeviceResourceSample `json:"mem"`
}

// DeviceResourceSample is one point in a resource usage time series.
type DeviceResourceSample struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// ProxyGet executes a GET request on a managed FortiGate through the FMG
// device proxy. The resource path is a FortiGate API path (e.g.
// "/api/v2/monitor/system/status"). Returns the raw JSON response data.
func (c *Client) ProxyGet(ctx context.Context, adom, device, resource string) (json.RawMessage, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) || !validName(device) {
		return nil, fmt.Errorf("%w: adom=%q device=%q", ErrInvalidName, adom, device)
	}

	target := fmt.Sprintf("adom/%s/device/%s", adom, device)
	return c.jsonExec(ctx, "sys/proxy/json", map[string]any{
		"target":  []string{target},
		"action":  "get",
		"resource": resource,
		"timeout": 20,
	})
}

// GetDeviceStatus retrieves system status from a managed FortiGate, including
// uptime, firmware version, serial number, and model.
func (c *Client) GetDeviceStatus(ctx context.Context, adom, device string) (*DeviceStatus, error) {
	data, err := c.ProxyGet(ctx, adom, device, "/api/v2/monitor/system/status")
	if err != nil {
		return nil, fmt.Errorf("fortimgr: get device status (device=%s): %w", device, err)
	}

	var envelope struct {
		Results DeviceStatus `json:"results"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		// Some versions return the status at the top level
		var status DeviceStatus
		if err2 := json.Unmarshal(data, &status); err2 != nil {
			return nil, fmt.Errorf("fortimgr: parse device status: %w", err)
		}
		return &status, nil
	}
	return &envelope.Results, nil
}

// GetDeviceUptime returns the uptime in seconds of a managed FortiGate.
// Convenience wrapper around GetDeviceStatus.
func (c *Client) GetDeviceUptime(ctx context.Context, adom, device string) (int64, error) {
	status, err := c.GetDeviceStatus(ctx, adom, device)
	if err != nil {
		return 0, err
	}
	return status.Uptime, nil
}

// GetDeviceResourceUsage retrieves CPU and memory usage from a managed FortiGate.
func (c *Client) GetDeviceResourceUsage(ctx context.Context, adom, device string) (*DeviceResourceUsage, error) {
	data, err := c.ProxyGet(ctx, adom, device, "/api/v2/monitor/system/resource/usage?scope=global")
	if err != nil {
		return nil, fmt.Errorf("fortimgr: get device resource usage (device=%s): %w", device, err)
	}

	var envelope struct {
		Results DeviceResourceUsage `json:"results"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("fortimgr: parse device resource usage: %w", err)
	}
	return &envelope.Results, nil
}

// ForwardGet reads a FMG-native API path via the forward endpoint.
// Useful for reading device config, dvmdb entries, and other FMG-stored data
// that doesn't require proxying to the managed FortiGate.
func (c *Client) ForwardGet(ctx context.Context, apiURL string) (json.RawMessage, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	return c.forward(ctx, apiURL)
}

// GetGlobalSetting reads a single FortiManager global CLI setting by name.
// Uses the forward endpoint to read /cli/global/system/global.
func (c *Client) GetGlobalSetting(ctx context.Context, field string) (json.RawMessage, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.forwardExtra(ctx, "/cli/global/system/global", map[string]any{
		"fields": []string{field},
	})
	if err != nil {
		return nil, fmt.Errorf("fortimgr: get global setting %q: %w", field, err)
	}
	return data, nil
}

// GetSaveLastHitInAdomDB checks whether the save-last-hit-in-adomdb setting
// is enabled on FortiManager. Returns "enable" or "disable".
func (c *Client) GetSaveLastHitInAdomDB(ctx context.Context) (string, error) {
	data, err := c.GetGlobalSetting(ctx, "save-last-hit-in-adomdb")
	if err != nil {
		return "", err
	}

	// Response can be an object with the field, or an array with one object
	var single struct {
		Value string `json:"save-last-hit-in-adomdb"`
	}
	if err := json.Unmarshal(data, &single); err == nil && single.Value != "" {
		return single.Value, nil
	}

	var arr []struct {
		Value string `json:"save-last-hit-in-adomdb"`
	}
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) > 0 {
		return arr[0].Value, nil
	}

	return "", fmt.Errorf("fortimgr: cannot parse save-last-hit-in-adomdb from %s", string(data))
}
