# Plan

Development roadmap for fortimgr.

## 🏗️ Phase 1: Core Client — ✅ Done

| Component | Description |
|-----------|-------------|
| HTTP client | TLS configuration, cookie jar |
| Session management | CSRF token extraction and injection |
| Login / Logout | Explicit session lifecycle |
| Request envelope | FlatUI forward endpoint wrapper |
| Response unwrapping | Status code validation, error extraction |
| Functional options | `WithCredentials`, `WithInsecureTLS`, `WithTimeout`, `WithTransport` |

## 📦 Phase 2: Resource Methods — ✅ Done

ADOM-scoped read operations:

| Method | Description |
|--------|-------------|
| `ListDevices(ctx, adom)` | Managed FortiGate devices |
| `ListPolicyPackages(ctx, adom)` | Firewall policy packages |
| `ListPolicies(ctx, adom, pkgName)` | Firewall rules per package |
| `ListAddresses(ctx, adom)` | Address objects (ipmask, iprange, fqdn, geography, wildcard) |
| `ListAddressGroups(ctx, adom)` | Address group objects |
| `ListServices(ctx, adom)` | Custom service definitions |
| `ListServiceGroups(ctx, adom)` | Service group objects |
| `ListSchedulesRecurring(ctx, adom)` | Recurring schedules |
| `ListSchedulesOnetime(ctx, adom)` | One-time schedules |
| `ListVirtualIPs(ctx, adom)` | Virtual IP / port forwarding objects |
| `ListIPPools(ctx, adom)` | NAT IP pool objects |

## ✅ Phase 3: Quality — ✅ Done

| Task | Description |
|------|-------------|
| Unit tests | httptest-based mock server, table-driven conversion tests |
| Smoke test | Live FortiManager test (env-var credentials, `go run smoke.go`) |
| GoDoc | Comments on all exported types and methods |

## 🔮 Phase 4: Extended Resources

Candidates based on JSON-RPC API coverage — add as needed:

| Resource | Category |
|----------|----------|
| Interfaces | Network |
| Routes / Routing tables | Network |
| VPN (IPsec, SSL) | VPN |
| System settings | System |
| HA status | System |
| ADOM management | System |

## ❌ Non-Goals

| Scope | Reason |
|-------|--------|
| Write operations (create/update/delete) | Read-only SDK |
| FortiGate direct API | Separate project |
| FortiAnalyzer API | Separate project |
| Full JSON-RPC parity | Exists for hardened environments only |
