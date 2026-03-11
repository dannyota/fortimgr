# Architecture

Go SDK for FortiManager's FlatUI API — the same HTTP API the web UI uses.

## 🔄 FlatUI Protocol

### Authentication

Session-based auth with CSRF token protection:

| Step | Endpoint | Details |
|------|----------|---------|
| Login | `POST /cgi-bin/module/flatui_auth` | Body: `{"secretkey": "<pw>", "logintype": 0}` |
| Token | Set by login response | `HTTP_CSRF_TOKEN` cookie |
| Requests | CSRF header (varies by endpoint) | See transport endpoints below |
| Logout | `POST /cgi-bin/module/flatui_auth` | Logout action |

### Transport Endpoints

FortiManager's FlatUI has two transport endpoints with different request/response formats:

#### Forward (`/cgi-bin/module/forward`)

Used for most API operations (device, policy, firewall objects, etc.).

```
POST /cgi-bin/module/forward
X-CSRFToken: <token>
```

```json
{
  "id": 1,
  "method": "get",
  "params": [{ "url": "/dvmdb/adom/root/device" }]
}
```

Response envelope:

```json
{
  "code": 0,
  "data": {
    "result": [{
      "status": { "code": 0, "message": "OK" },
      "data": [ ... ]
    }]
  }
}
```

| Field | Description |
|-------|-------------|
| `id` | Incremental request counter |
| `method` | Forwarded method (`"get"`) |
| `params[].url` | Internal FortiManager API path |
| `code` | Transport status (0 = success) |
| `result[].status.code` | API status (0 = success, -6 = session expired, -11 = no permission) |
| `result[].data` | Response payload |

#### Proxy (`/cgi-bin/module/flatui_proxy`)

Used for GUI-specific operations (system status, firmware management).

```
POST /cgi-bin/module/flatui_proxy
xsrf-token: <token>
```

```json
{
  "url": "/gui/sys/config",
  "method": "get"
}
```

Response envelope (no outer `code`/`data` wrapper):

```json
{
  "result": [{
    "status": { "code": 0, "message": "OK" },
    "data": { ... }
  }]
}
```

| Difference | Forward | Proxy |
|------------|---------|-------|
| Endpoint | `/cgi-bin/module/forward` | `/cgi-bin/module/flatui_proxy` |
| CSRF header | `X-CSRFToken` | `xsrf-token` |
| Request format | `{id, method, params: [{url}]}` | `{url, method}` |
| Response wrapper | `{code, data: {result: [...]}}` | `{result: [...]}` |
| Custom methods | Always `"get"` | Varies (e.g. `"loadFirmwareDataGroupByVersion"`) |

### Status Codes

Both endpoints share the same status codes in `result[].status.code`:

| Code | Meaning | SDK Error |
|------|---------|-----------|
| 0 | Success | — |
| -3 | Object does not exist | `APIError` |
| -6 | Session expired | `ErrSessionExpired` (auto-relogin) |
| -11 | No permission | `ErrPermission` |

## 📂 Package Layout

```
danny.vn/fortimgr/
├── client.go             # Client, NewClient, Login, Logout, Close
├── option.go             # WithCredentials, WithInsecureTLS, WithTimeout, etc.
├── request.go            # Forward + proxy transports, generic get[T]
├── response.go           # Forward response unwrapping, error extraction
├── errors.go             # ErrAuth, ErrPermission, ErrCertificate, etc.
├── types.go              # All public domain types (29 types)
├── convert.go            # Enum maps, subnet/IP/schedule formatting helpers
│
├── adom.go               # ListADOMs
├── device.go             # ListDevices
├── policy.go             # ListPolicyPackages, ListPolicies
├── address.go            # ListAddresses, ListAddressGroups
├── service.go            # ListServices, ListServiceGroups
├── schedule.go           # ListSchedulesRecurring, ListSchedulesOnetime
├── virtualip.go          # ListVirtualIPs
├── ippool.go             # ListIPPools
├── zone.go               # ListZones
├── vdom.go               # ListVDOMs
├── interface.go          # ListInterfaces
├── route.go              # ListStaticRoutes
├── secprofile.go         # AV, IPS, WebFilter, AppControl, SSL/SSH profiles
├── user.go               # Users, UserGroups, LDAP, RADIUS
├── vpn.go                # IPSecPhase1, IPSecPhase2
├── system.go             # SystemStatus (proxy)
├── firmware.go           # ListDeviceFirmware (proxy)
│
├── testhelper_test.go    # Shared httptest server (forward + proxy)
├── *_test.go             # Unit tests for all resources
└── smoke.go              # Live FortiManager smoke test (go run)
```

## 🏛️ Design Decisions

| Decision | Rationale |
|----------|-----------|
| Flat package | ~33 methods, single concern, no sub-packages needed |
| Functional options | Clean constructor, extensible (rate limiting, custom transports) |
| Explicit Login/Logout | No hidden network calls, visible session lifecycle |
| Read-only | Inventory/audit use case only, write operations too risky on undocumented API |
| ADOM parameter | Most resources scoped to Administrative Domain (default: `"root"`) |
| Device parameter | VDOMs, interfaces, routes scoped to specific device |
| Generic `get[T]` | Type-safe unmarshalling, eliminates boilerplate across resource methods |
| Separate DTOs | API structs stay unexported; public types have clean field names |
| Dual transport | Forward for `/dvmdb/`, `/pm/` paths; Proxy for `/gui/` paths |
| Auto-relogin | Both transports retry once on session expiry (code -6) |

## ⚠️ TLS

| Issue | Solution |
|-------|----------|
| Self-signed certs | `WithInsecureTLS()` |
| Negative X.509 serial numbers (non-RFC 5280) | `WithX509NegativeSerial()` |
