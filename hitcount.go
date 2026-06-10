package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// PolicyHitCount holds per-policy hit count statistics retrieved from a
// managed FortiGate via the FortiManager hit count task.
type PolicyHitCount struct {
	PolicyID     int    `json:"policyid"`
	Name         string `json:"name"`
	HitCount     int64  `json:"hitcount"`
	FirstHit     int64  `json:"first_hit"`
	LastHit      int64  `json:"last_hit"`
	Bytes        int64  `json:"byte"`
	Packets      int64  `json:"pkts"`
	SessionCount int    `json:"sesscount"`
	FirstSession int64  `json:"first_session"`
	LastSession  int64  `json:"last_session"`
	SrcIntf      string `json:"srcintf"`
	DstIntf      string `json:"dstintf"`
	UUID         string `json:"uuid"`
}

// TaskStatus represents the state of a FortiManager background task.
type TaskStatus struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	State   int    `json:"state"`
	Percent int    `json:"percent"`
	NumDone int    `json:"num_done"`
	NumErr  int    `json:"num_err"`
}

// RefreshHitCounts triggers FortiManager to collect hit counts from the
// managed FortiGate for one policy package. Returns the task ID to poll.
func (c *Client) RefreshHitCounts(ctx context.Context, adom string, adomOID, pkgOID int) (int, error) {
	if !c.LoggedIn() {
		return 0, ErrNotLoggedIn
	}

	data, err := c.jsonExec(ctx, "sys/hitcount", map[string]any{
		"adom":     adom,
		"adom_oid": adomOID,
		"pkg_oid":  pkgOID,
	})
	if err != nil {
		return 0, fmt.Errorf("fortimgr: trigger hit count refresh: %w", err)
	}

	var result struct {
		Task int `json:"task"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, fmt.Errorf("fortimgr: parse hit count task id: %w", err)
	}
	if result.Task == 0 {
		return 0, fmt.Errorf("fortimgr: hit count refresh returned no task id")
	}
	return result.Task, nil
}

// PollTask polls a FortiManager background task until it completes (state=4)
// or the context is cancelled. Polls every second.
func (c *Client) PollTask(ctx context.Context, taskID int) (*TaskStatus, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	url := fmt.Sprintf("/task/task/%d", taskID)
	for {
		data, err := c.forward(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fortimgr: poll task %d: %w", taskID, err)
		}

		var status TaskStatus
		if err := json.Unmarshal(data, &status); err != nil {
			return nil, fmt.Errorf("fortimgr: parse task status: %w", err)
		}

		switch status.State {
		case 4: // done
			if status.NumErr > 0 {
				return &status, fmt.Errorf("fortimgr: task %d completed with %d errors", taskID, status.NumErr)
			}
			return &status, nil
		case 3: // error
			return &status, fmt.Errorf("fortimgr: task %d failed", taskID)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// GetHitCountResults retrieves per-policy hit counts from a completed
// hit count refresh task.
func (c *Client) GetHitCountResults(ctx context.Context, taskID int) ([]PolicyHitCount, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.jsonExec(ctx, "sys/task/result", map[string]any{
		"taskid": taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("fortimgr: get hit count results: %w", err)
	}

	var envelope struct {
		FirewallPolicy []PolicyHitCount `json:"firewall policy"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("fortimgr: parse hit count results: %w", err)
	}
	return envelope.FirewallPolicy, nil
}

// waitForStableResults re-reads the hit-count task result until the row count
// stops growing. The FMG hit-count task dispatches sub-requests to managed
// FortiGates; state=4 means the task container finished dispatching, but
// device responses may still be accumulating in the result set. Two
// consecutive reads with the same count (1s apart) confirm the result is
// stable. At most maxStableWaits extra reads are made (then the current
// result is returned as-is).
func (c *Client) waitForStableResults(ctx context.Context, taskID int, initial []PolicyHitCount) ([]PolicyHitCount, error) {
	const maxStableWaits = 10

	prev := len(initial)
	results := initial
	for i := 0; i < maxStableWaits; i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-time.After(time.Second):
		}

		latest, err := c.GetHitCountResults(ctx, taskID)
		if err != nil {
			return results, err
		}
		if len(latest) == prev {
			return latest, nil
		}
		results = latest
		prev = len(latest)
	}
	return results, nil
}

// ListPolicyHitCounts is a convenience method that triggers a hit count
// refresh for one policy package, polls until complete, waits for the result
// set to stabilize (device responses may still arrive after state=4), and
// returns the per-policy hit count results.
func (c *Client) ListPolicyHitCounts(ctx context.Context, adom string, adomOID, pkgOID int) ([]PolicyHitCount, error) {
	taskID, err := c.RefreshHitCounts(ctx, adom, adomOID, pkgOID)
	if err != nil {
		return nil, err
	}

	if _, err := c.PollTask(ctx, taskID); err != nil {
		return nil, err
	}

	initial, err := c.GetHitCountResults(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return c.waitForStableResults(ctx, taskID, initial)
}

// PackageHitCounts holds per-policy hit counts for one policy package.
type PackageHitCounts struct {
	PackageName string
	PackageOID  int
	Policies    []PolicyHitCount
}

// ListAllPolicyHitCounts triggers hit count refresh for multiple packages in
// parallel (all triggers first, then poll all, then collect and stabilize all
// results). Errors on individual packages are collected and returned as a
// joined error; successfully collected packages are always returned so the
// caller can use partial results if it chooses.
func (c *Client) ListAllPolicyHitCounts(ctx context.Context, adom string, adomOID int, packages []PolicyPackage) ([]PackageHitCounts, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	type taskRef struct {
		pkg    PolicyPackage
		taskID int
	}

	// Step 1: trigger all refreshes.
	var tasks []taskRef
	var errs []error
	for _, pkg := range packages {
		taskID, err := c.RefreshHitCounts(ctx, adom, adomOID, pkg.OID)
		if err != nil {
			errs = append(errs, fmt.Errorf("package %q (OID %d): trigger: %w", pkg.Name, pkg.OID, err))
			continue
		}
		tasks = append(tasks, taskRef{pkg: pkg, taskID: taskID})
	}

	// Step 2: poll, collect, and stabilize each task's results.
	var results []PackageHitCounts
	for _, t := range tasks {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		if _, err := c.PollTask(ctx, t.taskID); err != nil {
			errs = append(errs, fmt.Errorf("package %q (OID %d): poll: %w", t.pkg.Name, t.pkg.OID, err))
			continue
		}
		initial, err := c.GetHitCountResults(ctx, t.taskID)
		if err != nil {
			errs = append(errs, fmt.Errorf("package %q (OID %d): results: %w", t.pkg.Name, t.pkg.OID, err))
			continue
		}
		hits, err := c.waitForStableResults(ctx, t.taskID, initial)
		if err != nil {
			errs = append(errs, fmt.Errorf("package %q (OID %d): stabilize: %w", t.pkg.Name, t.pkg.OID, err))
			continue
		}
		results = append(results, PackageHitCounts{
			PackageName: t.pkg.Name,
			PackageOID:  t.pkg.OID,
			Policies:    hits,
		})
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("fortimgr: %d/%d packages had errors: %w",
			len(errs), len(packages), joinErrors(errs))
	}
	return results, nil
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msg := errs[0].Error()
	for _, e := range errs[1:] {
		msg += "; " + e.Error()
	}
	return fmt.Errorf("%s", msg)
}

// jsonExec sends an exec request via /cgi-bin/module/flatui/json.
// This endpoint uses the same session and xsrf-token as flatui_proxy but
// accepts the forward-style method+params envelope with exec support.
func (c *Client) jsonExec(ctx context.Context, apiURL string, execData map[string]any) (json.RawMessage, error) {
	data, err := c.doJSONExec(ctx, apiURL, execData)
	if err == ErrSessionExpired {
		if loginErr := c.Login(ctx); loginErr != nil {
			return nil, fmt.Errorf("fortimgr: re-login after session expired: %w", loginErr)
		}
		return c.doJSONExec(ctx, apiURL, execData)
	}
	return data, err
}

func (c *Client) doJSONExec(ctx context.Context, apiURL string, execData map[string]any) (json.RawMessage, error) {
	param := map[string]any{"url": apiURL}
	if execData != nil {
		param["data"] = execData
	}
	payload := map[string]any{
		"id":     atomic.AddInt64(&c.requestID, 1),
		"method": "exec",
		"params": []map[string]any{param},
	}

	// flatui/json returns {"code":N, "data":{"result":[...]}} — the flatUIResponse shape.
	var result flatUIResponse
	if err := c.postModule(ctx, "/cgi-bin/module/flatui/json", "xsrf-token", payload, &result); err != nil {
		return nil, err
	}
	return checkResponse(&result)
}
