# Architecture

Go SDK for FortiManager's FlatUI API — the same HTTP API the web UI uses.

## 🔄 FlatUI Protocol

### Authentication

Session-based auth with CSRF token protection:

| Step | Endpoint | Details |
|------|----------|---------|
| Login | `POST /cgi-bin/module/flatui_auth` | Body: `{"secretkey": "<pw>", "logintype": 0}` |
| Token | Set by login response | `HTTP_CSRF_TOKEN` cookie |
| Requests | `X-CSRFToken` header | Extracted from cookie |
| Logout | `POST /cgi-bin/module/flatui_auth` | Logout action |

### Request Envelope

All data requests go through a single forwarding endpoint:

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

| Field | Description |
|-------|-------------|
| `id` | Incremental request counter |
| `method` | Forwarded method (`"get"`) |
| `params[].url` | Internal FortiManager API path |

### Response Envelope

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
| `code` | Transport status (0 = success) |
| `result[].status.code` | API status (0 = success, -11 = no permission) |
| `result[].data` | Response payload |

## 📂 Package Layout

```
danny.vn/fortimgr/
├── client.go       # Client, NewClient, Login, Logout, Close
├── option.go       # WithCredentials, WithInsecureTLS, WithTimeout, WithTransport
├── request.go      # FlatUI request envelope, ID counter, forward method
├── response.go     # Response unwrapping, error extraction
├── errors.go       # ErrAuth, ErrPermission, ErrCertificate
├── device.go       # ListDevices
├── policy.go       # ListPolicyPackages, ListPolicies
├── address.go      # ListAddresses, ListAddressGroups
├── service.go      # ListServices, ListServiceGroups
├── schedule.go     # ListSchedulesRecurring, ListSchedulesOnetime
├── vip.go          # ListVIPs
├── ippool.go       # ListIPPools
├── types.go        # Device, Policy, Address, etc.
└── convert.go      # Subnet/IP formatting helpers
```

## 🏛️ Design Decisions

| Decision | Rationale |
|----------|-----------|
| Flat package | ~11 methods, no sub-packages needed |
| Functional options | Clean constructor, extensible (rate limiting, custom transports) |
| Explicit Login/Logout | No hidden network calls, visible session lifecycle |
| Read-only | Inventory/audit use case only, write operations too risky on undocumented API |
| ADOM parameter | All resources scoped to Administrative Domain (default: `"root"`) |

## ⚠️ TLS

| Issue | Solution |
|-------|----------|
| Self-signed certs | `WithInsecureTLS()` |
| Negative X.509 serial numbers (non-RFC 5280) | `GODEBUG=x509negativeserial=1` |
