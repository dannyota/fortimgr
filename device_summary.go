package fortimgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var deviceSummaryHAModes = map[string]string{
	"0": "standalone",
	"1": "active-passive",
	"2": "active-active",
}

var deviceSummaryHAUpgradeModes = map[string]string{
	"0": "unset",
	"1": "normal",
	"2": "uninterruptible",
}

var deviceSummaryInstallPattern = regexp.MustCompile(`\(([^)]+)\).*Installed By:\s*(.+)$`)

// DeviceSummary retrieves the read-only install/configuration summary shown by
// FortiManager's device dashboard for a managed device.
//
// The FlatUI endpoint returns one aggregate object for the named device and is
// not paginated.
func (c *Client) DeviceSummary(ctx context.Context, adom, device string) (*DeviceSummary, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) || !validName(device) {
		return nil, fmt.Errorf("%w: adom=%q device=%q", ErrInvalidName, adom, device)
	}

	data, err := c.proxyParams(ctx, "/gui/adom/dvm/device/summary", "get", map[string]string{"name": device})
	if err != nil {
		return nil, err
	}

	var root any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal device summary: %w", err)
	}

	node := findDeviceSummaryNode(root, device)
	if node == nil {
		return &DeviceSummary{ADOM: adom, Device: device}, nil
	}
	summary := convertDeviceSummary(adom, device, node)
	return &summary, nil
}

func findDeviceSummaryNode(v any, device string) map[string]any {
	switch val := v.(type) {
	case map[string]any:
		if child, ok := val[device]; ok {
			if m, ok := child.(map[string]any); ok {
				return m
			}
		}
		if matchesDeviceSummaryNode(val, device) || looksLikeDeviceSummaryRoot(val) {
			return val
		}
		for _, child := range val {
			if found := findDeviceSummaryNode(child, device); found != nil {
				return found
			}
		}
	case []any:
		for _, child := range val {
			if found := findDeviceSummaryNode(child, device); found != nil {
				return found
			}
		}
	}
	return nil
}

func looksLikeDeviceSummaryRoot(m map[string]any) bool {
	_, hasSysConfig := m["sysConfig"]
	_, hasSysInfo := m["sysInfo"]
	return hasSysConfig || hasSysInfo
}

func matchesDeviceSummaryNode(m map[string]any, device string) bool {
	for _, key := range []string{"name", "device", "dev_name", "devName", "hostname"} {
		if strings.EqualFold(toString(m[key]), device) {
			return true
		}
	}
	return false
}

func convertDeviceSummary(adom, device string, m map[string]any) DeviceSummary {
	lastInstallation := firstNonEmpty(
		toString(pathValue(m, "sysConfig", "installTracking", "lastInstallation", "value")),
		toString(lookupSummaryValue(m, "last_installation", "lastInstallation")),
	)
	installTime, installedBy := parseSummaryInstallation(lastInstallation)
	if installTime.IsZero() {
		installTime = summaryTime(lookupSummaryValue(m, "last_install_time", "last_installed_time", "lastInstallTime", "install_time"))
	}
	if installedBy == "" {
		installedBy = toString(lookupSummaryValue(m, "installed_by", "install_user", "installUser", "user"))
	}

	s := DeviceSummary{
		ADOM:         adom,
		Device:       firstNonEmpty(toString(lookupSummaryValue(m, "name", "device", "dev_name", "devName")), device),
		Hostname:     toString(pathValue(m, "sysInfo", "hostName", "value")),
		SerialNumber: toString(pathValue(m, "sysInfo", "sn", "value")),
		Firmware:     toString(pathValue(m, "sysInfo", "firmware", "value")),
		ConfigStatus: firstNonEmpty(
			toString(pathValue(m, "sysConfig", "syncStatus", "value")),
			toString(pathValue(m, "sysConfig", "installTracking", "confChgStatues", "value")),
			mapEnum(toString(lookupSummaryValue(m, "conf_status", "config_status", "configStatus", "db_status")), confStatuses),
		),
		TotalRevisions: firstNonZero(
			toInt(pathValue(m, "sysConfig", "revision", "value")),
			toInt(lookupSummaryValue(m, "total_revision", "total_revisions", "revision_total", "revision_count", "rev_count")),
		),
		LastInstalledRevision: firstNonZero(
			toInt(pathValue(m, "sysConfig", "installTracking", "lastInstallation", "revision")),
			toInt(lookupSummaryValue(m, "last_install_rev", "last_installed_revision", "last_install_revision", "lastInstallRevision")),
		),
		LastInstallation: lastInstallation,
		LastInstallTime:  installTime,
		InstalledBy:      installedBy,
		HAMode: firstNonEmpty(
			toString(pathValue(m, "sysInfo", "haStatus", "value")),
			mapEnum(toString(lookupSummaryValue(m, "ha_mode", "haMode", "haModeStr")), deviceSummaryHAModes),
		),
		HAUpgradeMode: mapEnum(toString(lookupSummaryValue(m, "ha_upgrade_mode", "haUpgradeMode")), deviceSummaryHAUpgradeModes),
		HAClusterName: firstNonEmpty(toString(pathValue(m, "sysInfo", "haName", "value")), toString(lookupSummaryValue(m, "ha_cluster_name", "haClusterName", "cluster_name", "clusterName"))),
		HAClusterID:   toInt(lookupSummaryValue(m, "ha_cluster_id", "haClusterID", "cluster_id", "clusterId", "ha_cluster")),
	}
	s.HAMembers = summaryHAMembers(firstNonNil(pathValue(m, "sysInfo", "haMember", "value", "records"), lookupSummaryValue(m, "ha", "ha_slave", "ha_members", "haMembers", "cluster_members", "members")))
	return s
}

func pathValue(m map[string]any, path ...string) any {
	var cur any = m
	for _, key := range path {
		asMap, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = asMap[key]
		if cur == nil {
			return nil
		}
	}
	return cur
}

func lookupSummaryValue(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok && !isEmptySummaryValue(v) {
			return v
		}
	}
	for _, child := range m {
		childMap, ok := child.(map[string]any)
		if !ok {
			continue
		}
		if v := lookupSummaryValue(childMap, keys...); !isEmptySummaryValue(v) {
			return v
		}
	}
	return nil
}

func isEmptySummaryValue(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func summaryTime(v any) time.Time {
	if t := unixToTime(v); !t.IsZero() {
		return t
	}
	s := strings.TrimSpace(toString(v))
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func parseSummaryInstallation(s string) (time.Time, string) {
	matches := deviceSummaryInstallPattern.FindStringSubmatch(s)
	if len(matches) != 3 {
		return time.Time{}, ""
	}
	return summaryTime(matches[1]), strings.TrimSpace(matches[2])
}

func summaryHAMembers(v any) []DeviceSummaryHAMember {
	if wrapped, ok := v.(map[string]any); ok {
		v = firstNonNil(wrapped["records"], pathValue(wrapped, "value", "records"))
	}
	items, ok := v.([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	members := make([]DeviceSummaryHAMember, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		members = append(members, DeviceSummaryHAMember{
			OID:          toInt(lookupSummaryValue(m, "oid")),
			Name:         toString(lookupSummaryValue(m, "name", "hostname")),
			SerialNumber: toString(lookupSummaryValue(m, "sn", "serial", "serialNumber", "serial_number")),
			Role:         mapEnum(toString(lookupSummaryValue(m, "role", "ha_role", "haRole")), haRoles),
			Status:       toInt(lookupSummaryValue(m, "status")),
			SyncStatus:   mapEnum(toString(lookupSummaryValue(m, "sync_status", "syncStatus", "conf_status", "config_status")), confStatuses),
		})
	}
	if len(members) == 0 {
		return nil
	}
	return members
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if !isEmptySummaryValue(v) {
			return v
		}
	}
	return nil
}
