//go:build livecapture

package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type liveCaptureConfig struct {
	Address            string `json:"address"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	ADOM               string `json:"adom"`
	InsecureTLS        bool   `json:"insecure_tls"`
	X509NegativeSerial bool   `json:"x509_negative_serial"`
}

func TestLiveCaptureRawAndSDK(t *testing.T) {
	cfg := loadLiveCaptureConfig(t)

	var opts []ClientOption
	opts = append(opts, WithCredentials(cfg.Username, cfg.Password))
	if cfg.InsecureTLS {
		opts = append(opts, WithInsecureTLS())
	}
	if cfg.X509NegativeSerial {
		opts = append(opts, WithX509NegativeSerial())
	}

	ctx := context.Background()
	client, err := NewClient(cfg.Address, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Login(ctx); err != nil {
		t.Fatal(err)
	}

	adom := cfg.ADOM
	if adom == "" {
		adoms, err := client.ListADOMs(ctx)
		if err != nil {
			t.Fatal(err)
		}
		adom = firstEnabledADOM(adoms)
	}
	if adom == "" {
		t.Fatal("no ADOM available")
	}

	devices, err := client.ListDevices(ctx, adom)
	if err != nil {
		t.Fatal(err)
	}

	oid, err := client.adomOID(ctx, adom)
	if err != nil {
		t.Fatal(err)
	}
	assignedURL := fmt.Sprintf("/gui/adoms/%d/devices/assignedpkgs", oid)
	assignedRaw, err := client.proxy(ctx, assignedURL, "get")
	if err != nil {
		t.Fatal(err)
	}
	assignedSDK, err := client.ListDeviceAssignedPackages(ctx, adom)
	if err != nil {
		t.Fatal(err)
	}
	writeLiveCaptureJSON(t, "raw/flatui/device_assigned_packages.json", json.RawMessage(assignedRaw))
	writeLiveCaptureJSON(t, "sdk/device_assigned_packages.json", assignedSDK)

	var assignedMap map[string]json.RawMessage
	if err := json.Unmarshal(assignedRaw, &assignedMap); err != nil {
		t.Fatal(err)
	}
	if len(assignedMap) != len(assignedSDK) {
		t.Fatalf("assigned package count mismatch: raw=%d sdk=%d", len(assignedMap), len(assignedSDK))
	}

	if len(devices) == 0 {
		t.Log("no devices available; skipped device summary capture")
	} else {
		summaryRaw, err := client.proxyParams(ctx, "/gui/adom/dvm/device/summary", "get", map[string]string{"name": devices[0].Name})
		if err != nil {
			t.Fatal(err)
		}
		summarySDK, err := client.DeviceSummary(ctx, adom, devices[0].Name)
		if err != nil {
			t.Fatal(err)
		}
		writeLiveCaptureJSON(t, "raw/flatui/device_summary.json", json.RawMessage(summaryRaw))
		writeLiveCaptureJSON(t, "sdk/device_summary.json", summarySDK)
		if summarySDK == nil {
			t.Fatal("nil SDK device summary")
		}
		if summarySDK.Device == "" {
			t.Fatal("device summary missing Device")
		}

		captureDeviceNetworkRaw(t, ctx, client, devices[0].Name)
	}

	workflowSessions, err := client.ListWorkflowSessions(ctx, adom)
	if err != nil {
		t.Fatal(err)
	}
	writeLiveCaptureJSON(t, "sdk/workflow_sessions.json", workflowSessions)
	if len(workflowSessions) == 0 {
		t.Log("no workflow sessions available; skipped workflow log capture")
		return
	}
	workflowLogRaw, err := client.forward(ctx, fmt.Sprintf("/dvmdb/adom/%s/workflow/%d/wflog", adom, workflowSessions[0].SessionID))
	if err != nil {
		t.Fatal(err)
	}
	workflowLogs, err := client.ListWorkflowLogs(ctx, adom, workflowSessions[0].SessionID)
	if err != nil {
		t.Fatal(err)
	}
	writeLiveCaptureJSON(t, "raw/forward/workflow_log.json", json.RawMessage(workflowLogRaw))
	writeLiveCaptureJSON(t, "sdk/workflow_log.json", workflowLogs)
	var workflowLogItems []json.RawMessage
	if err := json.Unmarshal(workflowLogRaw, &workflowLogItems); err != nil {
		t.Fatal(err)
	}
	if len(workflowLogItems) != len(workflowLogs) {
		t.Fatalf("workflow log count mismatch: raw=%d sdk=%d", len(workflowLogItems), len(workflowLogs))
	}

	firmwareRaw, err := client.proxy(ctx, "/gui/adom/dvm/firmware/management", "loadFirmwareDataGroupByVersion")
	if err != nil {
		t.Logf("skip firmware raw capture: %v", err)
	} else {
		firmwareSDK, err := client.ListDeviceFirmware(ctx)
		if err != nil {
			t.Fatal(err)
		}
		writeLiveCaptureJSON(t, "raw/flatui/device_firmware.json", json.RawMessage(firmwareRaw))
		writeLiveCaptureJSON(t, "sdk/device_firmware.json", firmwareSDK)
	}

	updatePathRaw, err := client.proxy(ctx, "/gui/adom/dvm/device/firmware", "getUpdatePath")
	if err != nil {
		t.Logf("skip firmware update path capture: %v", err)
	} else {
		updatePathSDK, err := client.ListFirmwareUpgradePaths(ctx)
		if err != nil {
			t.Fatal(err)
		}
		writeLiveCaptureJSON(t, "raw/flatui/firmware_update_path.json", json.RawMessage(updatePathRaw))
		writeLiveCaptureJSON(t, "sdk/firmware_update_path.json", updatePathSDK)
		var updatePathItems []string
		if err := json.Unmarshal(updatePathRaw, &updatePathItems); err != nil {
			t.Fatal(err)
		}
		if len(updatePathItems) != len(updatePathSDK) {
			t.Fatalf("firmware update path count mismatch: raw=%d sdk=%d", len(updatePathItems), len(updatePathSDK))
		}
	}

	psirtRaw, err := client.proxy(ctx, fmt.Sprintf("/gui/adoms/%d/dvm/psirt", oid), "get")
	if err != nil {
		t.Logf("skip PSIRT capture: %v", err)
	} else {
		psirtSDK, err := client.DevicePSIRT(ctx, adom)
		if err != nil {
			t.Fatal(err)
		}
		writeLiveCaptureJSON(t, "raw/flatui/device_psirt.json", json.RawMessage(psirtRaw))
		writeLiveCaptureJSON(t, "sdk/device_psirt.json", psirtSDK)
		var psirtItems apiDevicePSIRTReport
		if err := json.Unmarshal(psirtRaw, &psirtItems); err != nil {
			t.Fatal(err)
		}
		if len(psirtItems.ByIRNumber) != len(psirtSDK.ByIRNumber) {
			t.Fatalf("PSIRT advisory count mismatch: raw=%d sdk=%d", len(psirtItems.ByIRNumber), len(psirtSDK.ByIRNumber))
		}
	}
}

func captureDeviceNetworkRaw(t *testing.T, ctx context.Context, client *Client, device string) {
	t.Helper()

	ifaces, err := client.ListInterfaces(ctx, device, "")
	if err != nil {
		t.Logf("skip device network capture: list interfaces: %v", err)
		return
	}
	vdom := firstInterfaceVDOM(ifaces)
	if vdom == "" {
		t.Log("skip device network capture: no interface VDOM found")
		return
	}

	captures := []struct {
		path string
		url  string
	}{
		{"raw/forward/device_dns.json", fmt.Sprintf("/pm/config/device/%s/global/system/dns", device)},
		{"raw/forward/device_ddns.json", fmt.Sprintf("/pm/config/device/%s/global/system/ddns", device)},
		{"raw/forward/device_ipam.json", fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/ipam", device, vdom)},
		{"raw/forward/device_sdwan.json", fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan", device, vdom)},
		{"raw/forward/device_sdwan_members.json", fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/members", device, vdom)},
		{"raw/forward/device_sdwan_services.json", fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/service", device, vdom)},
		{"raw/forward/device_sdwan_duplication.json", fmt.Sprintf("/pm/config/device/%s/vdom/%s/system/sdwan/duplication", device, vdom)},
	}
	for _, capture := range captures {
		data, err := client.forward(ctx, capture.url)
		if err != nil {
			t.Logf("skip %s: %v", capture.path, err)
			continue
		}
		writeLiveCaptureJSON(t, capture.path, json.RawMessage(data))
	}
}

func firstInterfaceVDOM(ifaces []Interface) string {
	for _, iface := range ifaces {
		if iface.VDOM != "" {
			return iface.VDOM
		}
	}
	return ""
}

func loadLiveCaptureConfig(t *testing.T) liveCaptureConfig {
	t.Helper()

	var cfg liveCaptureConfig
	if data, err := os.ReadFile(".fortimgr.json"); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("invalid .fortimgr.json: %v", err)
		}
	}
	if v := os.Getenv("FORTIMGR_ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("FORTIMGR_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("FORTIMGR_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("FORTIMGR_ADOM"); v != "" {
		cfg.ADOM = v
	}
	if cfg.Address == "" || cfg.Username == "" || cfg.Password == "" {
		t.Fatal("missing live FortiManager config")
	}
	return cfg
}

func firstEnabledADOM(adoms []ADOM) string {
	for _, a := range adoms {
		if a.Name == "root" && a.State == "enabled" {
			return a.Name
		}
	}
	for _, a := range adoms {
		if a.State == "enabled" {
			return a.Name
		}
	}
	if len(adoms) > 0 {
		return adoms[0].Name
	}
	return ""
}

func writeLiveCaptureJSON(t *testing.T, path string, v any) {
	t.Helper()
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	path = filepath.Join("samples", path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
