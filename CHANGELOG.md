# Changelog

## v1.0.1

Bug fixes discovered while running the SDK against a live restricted-admin session.

### Fixes

- **`ListADOMs()`** now filters to the session's accessible ADOMs by default. Previously it returned every ADOM on the box (including factory presets like `FortiAnalyzer`, `FortiMail`, `FortiWeb`, etc.), which caused restricted admins to see ADOMs they had no scope for and fail on subsequent calls. Pass `ListADOMs(ctx, true)` to retain the global view for superadmin tooling. Uses `/gui/sys/config` to resolve the session scope (same endpoint as `SystemStatus`).
- **`ListInterfaces(ctx, device, "")`** (or `"global"`) now uses the device-wide path `/pm/config/device/<dev>/global/system/interface`, returning every interface across all VDOMs in one call with each carrying its own `vdom` field. Restricted admins cannot read `/dvmdb/device/<dev>/vdom` to enumerate VDOMs first, so this is the only viable path for them. Callers can derive the VDOM set from the returned list. Passing a specific vdom still routes to the per-VDOM path.
- **`ListIPSecPhase1()` / `ListIPSecPhase2()`** corrected URL segment from `vpn.ipsec` (dotted) to `vpn/ipsec` (slash). The dotted form returned `-3 Object does not exist` on every FMG; the slash form is the valid path.

### Smoke test

- Rewrote the device loop to skip `ListVDOMs` (permission-denied for restricted admins) and instead fetch the global interface list per device, then derive the VDOM set from the result and call `ListStaticRoutes` once per derived VDOM.
- Added an `ADOMs: N accessible / M total` diagnostic line that calls both `ListADOMs()` and `ListADOMs(ctx, true)` so the filtered/global comparison is visible at a glance.

## v1.0.0

First stable release. Read-only Go SDK for FortiManager's FlatUI API.

### Core

- Dual transport: `forward` for config/device endpoints, `proxy` for system endpoints
- Session management with cookie jar and CSRF token handling
- Auto-relogin on session expiry (status code -6) for both transports
- Functional options: `WithCredentials`, `WithInsecureTLS`, `WithTimeout`, `WithTransport`, `WithHTTPClient`, `WithUserAgent`, `WithX509NegativeSerial`
- Sentinel errors: `ErrAuth`, `ErrNotLoggedIn`, `ErrPermission`, `ErrCertificate`, `ErrSessionExpired`, `ErrInvalidName`
- Input validation via `validName()` to prevent path injection
- Zero external dependencies (stdlib only)

### Resource Methods (29 total)

**System (proxy transport)**
- `SystemStatus()` — hostname, version, serial number, HA mode, platform
- `ListDeviceFirmware()` — firmware status for all managed devices

**Device Management**
- `ListADOMs()` — administrative domains
- `ListDevices(adom)` — managed FortiGate devices with firmware, HA, and status
- `ListVDOMs(device)` — virtual domains on a device
- `ListInterfaces(device, vdom)` — network interfaces with role, mode, allow access, VLAN
- `ListStaticRoutes(device, vdom)` — static routing entries

**Firewall Policy**
- `ListPolicyPackages(adom)` — policy packages with scope assignments
- `ListPolicies(adom, pkg)` — firewall rules per package

**Firewall Objects**
- `ListAddresses(adom)` — address objects (ipmask, iprange, fqdn, geography, wildcard)
- `ListAddressGroups(adom)` — address groups
- `ListServices(adom)` — custom service definitions
- `ListServiceGroups(adom)` — service groups
- `ListVirtualIPs(adom)` — virtual IP / port forwarding
- `ListIPPools(adom)` — NAT IP pools
- `ListZones(adom)` — system zones with intrazone traffic setting

**Scheduling**
- `ListSchedulesRecurring(adom)` — recurring schedules
- `ListSchedulesOnetime(adom)` — one-time schedules

**Security Profiles**
- `ListAntivirusProfiles(adom)` — scan mode, feature set, logging options
- `ListIPSSensors(adom)` — extended log, botnet scanning, malicious URL blocking
- `ListWebFilterProfiles(adom)` — inspection mode, content logging, FTGD error logging
- `ListAppControlProfiles(adom)` — deep inspection, unknown/other app actions
- `ListSSLSSHProfiles(adom)` — cert mode, MAPI/RPC over HTTPS, ALPN support

**User & Authentication**
- `ListUsers(adom)` — local users
- `ListUserGroups(adom)` — user groups with member lists
- `ListLDAPServers(adom)` — LDAP server configurations
- `ListRADIUSServers(adom)` — RADIUS server configurations

**VPN**
- `ListIPSecPhase1(adom)` — IPSec Phase 1 interfaces
- `ListIPSecPhase2(adom)` — IPSec Phase 2 interfaces

### Type Conversion

- Handles FortiManager's inconsistent JSON types (string, int, float64, array)
- Enum mapping from numeric API values to named FortiOS strings
- Subnet formatting with automatic dotted-mask to CIDR conversion
- Host addresses (/32) rendered without prefix length
