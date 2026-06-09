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

// ListPolicyHitCounts is a convenience method that triggers a hit count
// refresh for one policy package, polls until complete, and returns the
// per-policy hit count results.
func (c *Client) ListPolicyHitCounts(ctx context.Context, adom string, adomOID, pkgOID int) ([]PolicyHitCount, error) {
	taskID, err := c.RefreshHitCounts(ctx, adom, adomOID, pkgOID)
	if err != nil {
		return nil, err
	}

	if _, err := c.PollTask(ctx, taskID); err != nil {
		return nil, err
	}

	return c.GetHitCountResults(ctx, taskID)
}

// PackageHitCounts holds per-policy hit counts for one policy package.
type PackageHitCounts struct {
	PackageName string
	PackageOID  int
	Policies    []PolicyHitCount
}

// ListAllPolicyHitCounts triggers hit count refresh for multiple packages in
// parallel (all triggers first, then poll all, then collect all results).
// This is much faster than calling ListPolicyHitCounts sequentially when
// there are multiple packages — wall-clock drops from N*10s to ~10s.
func (c *Client) ListAllPolicyHitCounts(ctx context.Context, adom string, adomOID int, packages []PolicyPackage) ([]PackageHitCounts, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	type taskRef struct {
		pkg    PolicyPackage
		taskID int
	}

	// Step 1: trigger all refreshes
	var tasks []taskRef
	for _, pkg := range packages {
		taskID, err := c.RefreshHitCounts(ctx, adom, adomOID, pkg.OID)
		if err != nil {
			continue
		}
		tasks = append(tasks, taskRef{pkg: pkg, taskID: taskID})
	}

	// Step 2: poll all tasks and collect results
	var results []PackageHitCounts
	for _, t := range tasks {
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		if _, err := c.PollTask(ctx, t.taskID); err != nil {
			continue
		}
		hits, err := c.GetHitCountResults(ctx, t.taskID)
		if err != nil {
			continue
		}
		results = append(results, PackageHitCounts{
			PackageName: t.pkg.Name,
			PackageOID:  t.pkg.OID,
			Policies:    hits,
		})
	}

	return results, nil
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
