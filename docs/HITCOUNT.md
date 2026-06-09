# Policy Hit Count API

## Summary

Add `RefreshHitCounts` + `GetHitCountResults` to the SDK to collect per-policy
hit count, last-hit timestamp, bytes, packets, and session counts from managed
FortiGates via FortiManager.

## API Flow (captured from FMG Web UI via Playwright)

### Step 1 — Trigger hit count retrieval

```
POST /cgi-bin/module/flatui/json
Header: xsrf-token: <csrf>

{
  "method": "exec",
  "params": [{
    "url": "sys/hitcount",
    "data": {
      "adom": "root",
      "adom_oid": 100,
      "pkg_oid": 200
    }
  }],
  "id": "..."
}
```

Response:
```json
{"code": 0, "data": {"result": [{"data": {"task": 42}, "status": {"code": 0}}]}}
```

Returns a **task ID**. FMG asynchronously queries the managed FortiGate.

### Step 2 — Poll task until complete

```
POST /cgi-bin/module/forward
Header: X-CSRFToken: <csrf>

{"method": "get", "params": [{"url": "/task/task/42"}]}
```

Poll until `state == 4` and `percent == 100`. Typical completion: 2–6 seconds.

### Step 3 — Read task results

```
POST /cgi-bin/module/flatui/json
Header: xsrf-token: <csrf>

{
  "method": "exec",
  "params": [{
    "url": "sys/task/result",
    "data": {"taskid": 42}
  }],
  "id": "..."
}
```

Response (per policy):
```json
{
  "policyid": 10,
  "name": "allow-web-servers",
  "hitcount": 584210,
  "first_hit": 1700000000,
  "last_hit": 1780000000,
  "byte": 12345678900,
  "pkts": 9876543,
  "sesscount": 0,
  "first_session": 1700000100,
  "last_session": 1780000100,
  "srcintf": "port1",
  "dstintf": "port2"
}
```

Results are under `data.result[0].data["firewall policy"]` (array).

## SDK Changes Required

### Transport: `flatui/json` endpoint

The SDK has `forward` (`/cgi-bin/module/forward`) and `flatui_proxy`
(`/cgi-bin/module/flatui_proxy`) but NOT `flatui/json` (`/cgi-bin/module/flatui/json`).

`flatui/json` uses:
- Same session cookies as the other endpoints
- `xsrf-token` header (same as `flatui_proxy`, NOT `X-CSRFToken` like `forward`)
- Same response envelope as `flatui_proxy` (proxyResponse format)
- Supports `method: "exec"` (not just `"get"`)

### New types

```go
type PolicyHitCount struct {
    PolicyID     int
    Name         string
    HitCount     int64
    FirstHit     int64  // unix epoch
    LastHit      int64  // unix epoch
    Bytes        int64
    Packets      int64
    SessionCount int
    FirstSession int64
    LastSession  int64
    SrcIntf      string
    DstIntf      string
}
```

### New methods

```go
// RefreshHitCounts triggers FMG to collect hit counts from the managed
// FortiGate for one policy package. Returns the task ID to poll.
func (c *Client) RefreshHitCounts(ctx context.Context, adom string, adomOID, pkgOID int) (taskID int, err error)

// PollTask polls a task until completion (state=4). Returns an error if
// the task fails. Polls every second up to a timeout.
func (c *Client) PollTask(ctx context.Context, taskID int) error

// GetHitCountResults retrieves the per-policy hit counts from a completed
// hit count refresh task.
func (c *Client) GetHitCountResults(ctx context.Context, taskID int) ([]PolicyHitCount, error)

// ListPolicyHitCounts is a convenience method that triggers a refresh,
// polls until complete, and returns the results.
func (c *Client) ListPolicyHitCounts(ctx context.Context, adom string, adomOID, pkgOID int) ([]PolicyHitCount, error)
```

### Getting OIDs

`adom_oid`: already available from `ListADOMs` (the `OID` field on `apiADOM`).
`pkg_oid`: need to add `OID int` to `PolicyPackage` / `apiPolicyPackage` —
the forward API returns `oid` on every object, we just don't read it today.

## Notes

- `save-last-hit-in-adomdb` is NOT required — the task result endpoint
  returns hit counts directly from the managed FortiGate, regardless of
  the ADOM DB setting. The ADOM DB setting only affects the `_hitcount` /
  `_last_hit` fields on the policy config read.
- Hit count refresh takes 2–6 seconds per package (async via managed FG).
- Read-only operation — no config changes.
