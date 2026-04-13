# Changelog

## v1.0.3

Downstream-requested improvements for the hotpot data warehouse — policy install status, richer device sync state, and per-member HA role. Fully backwards compatible: every v1.0.2 symbol stays intact, all additions go to the end of existing structs so positional literals keep compiling.

### Added

- **`ListPackageInstallStatus(ctx, adom, pkg string) ([]PackageInstallStatus, error)`** — new method hitting `/pm/config/adom/{adom}/_package/status`. Distinguishes package **assignment** (device on scope list) from actual **installation** (config pushed and running on the FortiGate). `pkg` is optional — empty returns every package in the ADOM, non-empty filters server-side via a `filter: ["pkg","==",pkg]` clause. `PackageInstallStatus` fields: `ADOM`, `Package`, `Device`, `VDOM`, `Status` (`"installed"` / `"modified"` / `"never"` / `"unknown"` / `"imported"`).
- **`Device` struct additions** — `Hostname`, `ConfStatus` (`"unknown"` / `"insync"` / `"modified"`), `DevStatus` (`"none"` / `"auto_updated"` / `"installed"` / 13 others), `LastChecked` (`time.Time`, zero when `last_checked==0`), `LastResync` (same), `HARole` (`""` / `"master"` / `"slave"`), `HAMembers` (`[]HAMember`). All populated from the existing `/dvmdb/adom/{adom}/device` response — no new API calls. `HARole` is derived by matching the device name against `ha_slave[]`.
- **`HAMembers` + `HAMember` type** — surface every HA cluster member (including the standby) that FortiManager knows about for a given device record. FortiManager models each HA cluster as a **single** top-level device entry with `Name`/`Hostname` set to the primary's hostname — `ListDevices` has never returned standbys as separate rows and still doesn't. `HAMembers` is the only place where passive members appear. Each entry carries `Name`, `SerialNumber`, `Role` (`"master"` / `"slave"`), `Status` (`"online"` / `"offline"`), and `ConfStatus`. Empty for standalone devices.
- **`getExtra[T]` internal helper** — private generic wrapper alongside `get[T]`; forwards a GET whose `params[0]` merges extra fields (`filter`, `option`, …) into the payload. Used by `ListPackageInstallStatus`; existing call sites of `get[T]` are untouched.
- **Enum maps** — `confStatuses`, `devStatuses`, `haRoles` with raw-int passthrough for unmapped values (forward-compatible with future FortiManager schema additions).
- **`unixToTime` helper** — converts `int`/`float64`/`string` Unix timestamps (including `nil` / `0`) to `time.Time`, returning the zero value for "never" semantics.

### Notes on `HAMode` (legacy field unchanged)

The existing `Device.HAMode` still maps the raw `ha_mode` int via the legacy `deviceHAModes` table (`"0": "standalone", "1": "master", "2": "slave"`) — semantically this conflates topology and role, but behavior is preserved for v1.0.x callers. New code should prefer `HARole` for the per-member role and treat `HAMode` as opaque until a future major version where `HAMode` is cleaned up to mean topology only (`"standalone"` / `"a-p"` / `"a-a"`).

### Known gaps (planned for v1.1.0)

The `/pm/config/adom/{adom}/_package/status` endpoint does **not** expose `RevisionDeployed`, `RevisionLatest`, `LastInstallTime`, `ModifyState`, or `PendingChanges` — the live FortiManager API only returns the aggregate `status` string. Callers that need those details should join against ADOM revision history (`ListADOMRevisions` — shipping in v1.1.0 along with workflow sessions, normalized interfaces, SDN connectors, ISDB, and traffic shapers).

## v1.0.2

Friendlier IPsec naming matching the FortiGate GUI. Fully backwards compatible — no v1.0.1 symbol was renamed or removed.

### Added

- **`IPSecTunnel`** — zero-cost type alias for `IPSecPhase1` (`type IPSecTunnel = IPSecPhase1`). Values are interchangeable with no conversion.
- **`ListIPSecTunnels(ctx, adom)`** — one-line wrapper around `ListIPSecPhase1`. Shares all logic via the type alias.
- **`IPSecSelector`** — new struct mirroring `IPSecPhase2` with one field rename: `Phase1Name` → `Tunnel`. Kept as a distinct type (not an alias) so renaming the field on `IPSecPhase2` isn't required and v1.0.1 callers of `.Phase1Name` keep compiling.
- **`ListIPSecSelectors(ctx, adom)`** — delegates to `ListIPSecPhase2` to reuse the HTTP/JSON/mapping path, then copies each result into an `IPSecSelector` with the renamed field. Single point of truth for the API call.

### Rationale

Phase 1 / Phase 2 are IKE-RFC terms that force users to translate in their heads. The FortiGate GUI calls them "tunnel" and "selector", so matching that makes call sites read naturally:

```go
// Before
phase1, _ := c.ListIPSecPhase1(ctx, "root")
phase2, _ := c.ListIPSecPhase2(ctx, "root")
for _, p := range phase2 { fmt.Println(p.Phase1Name) }

// After
tunnels, _   := c.ListIPSecTunnels(ctx, "root")
selectors, _ := c.ListIPSecSelectors(ctx, "root")
for _, s := range selectors { fmt.Println(s.Tunnel) }
```

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
