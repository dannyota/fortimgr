# Plan

Development roadmap for fortimgr.

## 🏗️ Phase 1: Core Client — ✅ Done

| Component | Description |
|-----------|-------------|
| HTTP client | TLS configuration, cookie jar |
| Session management | CSRF token extraction and injection |
| Login / Logout | Explicit session lifecycle |
| Forward transport | `/cgi-bin/module/forward` envelope wrapper |
| Proxy transport | `/cgi-bin/module/flatui_proxy` envelope wrapper |
| Response unwrapping | Status code validation, error extraction |
| Auto-relogin | Retry on session expiry (code -6) |
| Functional options | `WithCredentials`, `WithInsecureTLS`, `WithTimeout`, `WithTransport`, `WithX509NegativeSerial` |

## 📦 Phase 2: Resource Methods — ✅ Done

### ADOM-Scoped (via forward transport)

| Method | Description |
|--------|-------------|
| `ListADOMs()` | Administrative Domains |
| `ListDevices(adom)` | Managed FortiGate devices |
| `ListPolicyPackages(adom)` | Firewall policy packages |
| `ListPolicies(adom, pkg)` | Firewall rules per package |
| `ListAddresses(adom)` | Address objects (ipmask, iprange, fqdn, geography, wildcard) |
| `ListAddressGroups(adom)` | Address group objects |
| `ListServices(adom)` | Custom service definitions |
| `ListServiceGroups(adom)` | Service group objects |
| `ListSchedulesRecurring(adom)` | Recurring schedules |
| `ListSchedulesOnetime(adom)` | One-time schedules |
| `ListVirtualIPs(adom)` | Virtual IP / port forwarding objects |
| `ListIPPools(adom)` | NAT IP pool objects |
| `ListZones(adom)` | System zones (interface grouping) |
| `ListAntivirusProfiles(adom)` | Antivirus profiles |
| `ListIPSSensors(adom)` | IPS sensors |
| `ListWebFilterProfiles(adom)` | Web filter profiles |
| `ListAppControlProfiles(adom)` | Application control profiles |
| `ListSSLSSHProfiles(adom)` | SSL/SSH inspection profiles |
| `ListUsers(adom)` | Local users |
| `ListUserGroups(adom)` | User groups |
| `ListLDAPServers(adom)` | LDAP server configs |
| `ListRADIUSServers(adom)` | RADIUS server configs |
| `ListIPSecPhase1(adom)` | IPSec Phase 1 interfaces |
| `ListIPSecPhase2(adom)` | IPSec Phase 2 interfaces |

### Device-Scoped (via forward transport)

| Method | Description |
|--------|-------------|
| `ListVDOMs(device)` | Virtual Domains on a device |
| `ListInterfaces(device, vdom)` | Network interfaces |
| `ListStaticRoutes(device, vdom)` | Static routing entries |

### System (via proxy transport)

| Method | Description |
|--------|-------------|
| `SystemStatus()` | FortiManager hostname, version, serial number, HA mode |
| `ListDeviceFirmware()` | Firmware status for all managed devices |

## ✅ Phase 3: Quality — ✅ Done

| Task | Description |
|------|-------------|
| Unit tests | httptest-based mock server (forward + proxy), table-driven conversion tests |
| Smoke test | Live FortiManager test (env-var credentials, `go run smoke.go`) |
| GoDoc | Comments on all exported types and methods |
| Input validation | Path injection prevention via `validName()` |

## 🔮 Phase 4: Future

Candidates based on JSON-RPC API coverage — add as needed:

| Resource | Category |
|----------|----------|
| SSL VPN settings | VPN |
| HA status | System |
| Log fetch | Logging |
| Event alerts | Monitoring |

## ❌ Non-Goals

| Scope | Reason |
|-------|--------|
| Write operations (create/update/delete) | Read-only SDK |
| FortiGate direct API | Separate project |
| FortiAnalyzer API | Separate project |
| Full JSON-RPC parity | Exists for hardened environments only |
