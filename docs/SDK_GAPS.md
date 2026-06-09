# SDK Gaps Discovered (2026-06-09 Firewall Review Session)

Findings from Playwright-based API discovery on a FortiManager instance.
Each item is a concrete API call pattern captured from the FMG web UI.

## 1. Device System Proxy (`sys/proxy/json`)

The FMG can proxy API calls to managed FortiGates via `sys/proxy/json`. The dashboard
uses this to show device uptime, resource usage, VM info, traffic history, etc.

**Transport:** `POST /cgi-bin/module/flatui/json` with `method: "exec"`
**CSRF:** Uses `xsrf-token` header (same as other flatui/json calls) — BUT from browser
evaluate it returns "Not match CSRFToken". The UI's JavaScript framework may use a
per-request or per-device CSRF token that we haven't captured. Needs further investigation.

**Captured calls from the dashboard:**

```
// System resource usage (CPU, memory)
sys/proxy/json → /api/v2/monitor/system/resource/usage?scope=global
sys/proxy/json → /api/v2/monitor/system/resource/usage?scope=root

// VM information
sys/proxy/json → /api/v2/monitor/system/vm-information

// Traffic history (per interface)
sys/proxy/json → /api/v2/monitor/system/traffic-history/interface?interface=port1&time_period=hour&vdom=root

// FortiToken
sys/proxy/json → /api/v2/monitor/user/fortitoken/select?&vdom=root
```

**What we need:** `/api/v2/monitor/system/status` — returns uptime, firmware version, serial
number. This is the standard FortiGate API for device status. The dashboard shows "Uptime:
335 days 11 hours..." so it DOES get this data — we just haven't found the exact call yet.

**SDK addition needed:**
```go
func (c *Client) ProxyGet(ctx context.Context, adom, device, resource string) (json.RawMessage, error)
```

## 2. Device Uptime

Confirmed visible in the UI: Device Manager → click device → Dashboard: Summary →
System Information → "Uptime: 335 days 11 hours 49 minutes 47 seconds"

**Source:** Either from `sys/proxy/json → /api/v2/monitor/system/status` or from a
dashboard-specific widget endpoint. The device summary endpoint (`/gui/adom/dvm/device/summary`)
does NOT include uptime — it has sysTime, firmware, hardware, HA status, but no uptime field.

**For hotpot:** Would feed into the firewall report to show how long hit counters have been
accumulating. Current workaround: derive counter age from `first_hit` in policy hit count results.

## 3. `save-last-hit-in-adomdb` Setting

FMG global setting that controls whether hit count `_last_hit` values survive FortiGate reboots.

**Check:** CLI only — `config system global / get save-last-hit-in-adomdb / end`
No web GUI equivalent. No known FlatUI API to read `system global` settings.

**Possible API:** Try `/cli/global/system/global` via forward endpoint — FMG JSON-RPC
supports reading CLI objects. Not yet tested.

**For hotpot:** Determines whether hit count data is reliable across reboots. If enabled
(likely, given first_hit=2023 predates 335-day uptime), hit counts cover the full policy
lifetime. If disabled, they cover only since last reboot.

## 4. Policy Package OID Resolution

**Done in v1.4.0:** `PolicyPackage.OID` and `ADOM.OID` now exposed.

## 5. Policy Hit Counts via Task

**Done in v1.4.0:** `ListPolicyHitCounts(adom, adomOID, pkgOID)` — trigger → poll → collect.

## Resolution (2026-06-09)

### sys/proxy/json — WORKS but limited

The CSRF issue was a Playwright cookie-parsing bug. The Go SDK handles it correctly.
However, `/api/v2/monitor/system/status` is **not in this FMG's proxy allowlist**
(`fos_json_api.json`). Only specific endpoints are whitelisted:
- `/api/v2/monitor/system/resource/usage` — WORKS (CPU, memory, sessions)
- `/api/v2/monitor/system/vm-information` — WORKS
- `/api/v2/monitor/system/traffic-history/interface` — WORKS
- `/api/v2/monitor/system/status` — BLOCKED ("Can not find this resource")
- `/api/v2/monitor/system/firmware` — BLOCKED
- `/api/v2/monitor/system/time` — BLOCKED

The FMG web UI shows device uptime on Dashboard > Summary, but it likely uses a
server-rendered widget or an internal endpoint not available through the FlatUI API.

### Device uptime — workaround via first_hit

Uptime is not directly available through the API. Workaround:
- From hit count task results: `first_hit` gives the earliest traffic match timestamp
- From FortiGate direct: `FirstUsed` in `MonitorPolicyStats`
- Both give the effective counter start date, which equals reboot date when
  `save-last-hit-in-adomdb` is disabled (default)

### save-last-hit-in-adomdb — CLI only

`/cli/global/system/global` returns "no permission" for the web API user. This setting
is only readable via SSH/console CLI access.

### SDK additions made

1. `ProxyGet(adom, device, resource)` — works for whitelisted FG monitor endpoints
2. `GetDeviceResourceUsage(adom, device)` — CPU/memory via proxy
3. `GetDeviceStatus(adom, device)` — implemented but blocked by allowlist on this FMG
4. `GetSaveLastHitInAdomDB()` — implemented but blocked by user permissions
5. `GetDeviceUptime(adom, device)` — wrapper around GetDeviceStatus, same limitation
