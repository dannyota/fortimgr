package fortimgr

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
)

func TestListPolicyHitCounts(t *testing.T) {
	taskDone := `{"id":1,"title":"hit count","state":4,"percent":100,"num_done":1,"num_err":0}`
	hits := `{"firewall policy":[{"policyid":1,"name":"rule-a","hitcount":42},{"policyid":2,"name":"rule-b","hitcount":0}]}`

	client := newTestClientFull(t, map[string]string{
		"/task/task/1": taskDone,
	}, nil, map[string]func(map[string]any) string{
		"sys/hitcount":    func(map[string]any) string { return `{"task":1}` },
		"sys/task/result": func(map[string]any) string { return hits },
	})

	results, err := client.ListPolicyHitCounts(t.Context(), "root", 1, 100)
	if err != nil {
		t.Fatalf("ListPolicyHitCounts: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d policies, want 2", len(results))
	}
	if results[0].HitCount != 42 {
		t.Fatalf("policy 1 hit count = %d, want 42", results[0].HitCount)
	}
	if results[1].HitCount != 0 {
		t.Fatalf("policy 2 hit count = %d, want 0 (never-used must be present)", results[1].HitCount)
	}
}

func TestListPolicyHitCounts_StabilizesGrowingResults(t *testing.T) {
	taskDone := `{"id":1,"title":"hit count","state":4,"percent":100,"num_done":1,"num_err":0}`

	var readCount atomic.Int32
	client := newTestClientFull(t, map[string]string{
		"/task/task/1": taskDone,
	}, nil, map[string]func(map[string]any) string{
		"sys/hitcount": func(map[string]any) string { return `{"task":1}` },
		"sys/task/result": func(map[string]any) string {
			n := readCount.Add(1)
			switch {
			case n <= 1:
				return `{"firewall policy":[{"policyid":1,"name":"r1","hitcount":10}]}`
			case n <= 2:
				return `{"firewall policy":[{"policyid":1,"name":"r1","hitcount":10},{"policyid":2,"name":"r2","hitcount":0}]}`
			default:
				return `{"firewall policy":[{"policyid":1,"name":"r1","hitcount":10},{"policyid":2,"name":"r2","hitcount":0}]}`
			}
		},
	})

	results, err := client.ListPolicyHitCounts(t.Context(), "root", 1, 100)
	if err != nil {
		t.Fatalf("ListPolicyHitCounts: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d policies, want 2 (stabilization should have waited for the second row)", len(results))
	}
	if readCount.Load() < 3 {
		t.Fatalf("expected at least 3 result reads (initial + growing + stable), got %d", readCount.Load())
	}
}

func TestListAllPolicyHitCounts_PropagatesErrors(t *testing.T) {
	taskDone := `{"id":1,"title":"hit count","state":4,"percent":100,"num_done":1,"num_err":0}`
	hits := `{"firewall policy":[{"policyid":1,"name":"r1","hitcount":5}]}`

	var taskIDSeq atomic.Int32
	client := newTestClientFull(t, map[string]string{
		"/task/task/1": taskDone,
		"/task/task/2": `{"id":2,"title":"hit count","state":3,"percent":0,"num_done":0,"num_err":1}`,
	}, nil, map[string]func(map[string]any) string{
		"sys/hitcount": func(params map[string]any) string {
			id := taskIDSeq.Add(1)
			return fmt.Sprintf(`{"task":%d}`, id)
		},
		"sys/task/result": func(params map[string]any) string {
			return hits
		},
	})

	packages := []PolicyPackage{
		{Name: "pkg-ok", OID: 100},
		{Name: "pkg-fail", OID: 200},
	}

	results, err := client.ListAllPolicyHitCounts(t.Context(), "root", 1, packages)
	if err == nil {
		t.Fatal("expected an error for the failed package, got nil")
	}
	if len(results) != 1 {
		t.Fatalf("got %d package results, want 1 (successful package)", len(results))
	}
	if results[0].PackageName != "pkg-ok" {
		t.Fatalf("successful package = %q, want pkg-ok", results[0].PackageName)
	}
}

func TestListAllPolicyHitCounts_IncludesZeroHitPolicies(t *testing.T) {
	taskDone := `{"id":1,"title":"hit count","state":4,"percent":100,"num_done":1,"num_err":0}`
	hits := `{"firewall policy":[{"policyid":1,"name":"active","hitcount":100},{"policyid":2,"name":"never-used","hitcount":0},{"policyid":3,"name":"also-never","hitcount":0}]}`

	client := newTestClientFull(t, map[string]string{
		"/task/task/1": taskDone,
	}, nil, map[string]func(map[string]any) string{
		"sys/hitcount":    func(map[string]any) string { return `{"task":1}` },
		"sys/task/result": func(map[string]any) string { return hits },
	})

	packages := []PolicyPackage{{Name: "test-pkg", OID: 100}}
	results, err := client.ListAllPolicyHitCounts(t.Context(), "root", 1, packages)
	if err != nil {
		t.Fatalf("ListAllPolicyHitCounts: %v", err)
	}
	if len(results) != 1 || len(results[0].Policies) != 3 {
		t.Fatalf("got %d packages / %d policies, want 1/3", len(results), len(results[0].Policies))
	}

	zeroCount := 0
	for _, p := range results[0].Policies {
		if p.HitCount == 0 {
			zeroCount++
		}
	}
	if zeroCount != 2 {
		t.Fatalf("zero-hit policies = %d, want 2", zeroCount)
	}
}

func TestPollTask_CompletedWithErrors(t *testing.T) {
	fixtures := map[string]string{
		"/task/task/99": `{"id":99,"title":"fail","state":4,"percent":100,"num_done":2,"num_err":1}`,
	}
	client := newTestClient(t, fixtures)

	status, err := client.PollTask(t.Context(), 99)
	if err == nil {
		t.Fatal("expected error for task with num_err > 0")
	}
	if status == nil || status.NumErr != 1 {
		t.Fatalf("status.NumErr = %v, want 1", status)
	}
}

func TestPollTask_FailedState(t *testing.T) {
	fixtures := map[string]string{
		"/task/task/99": `{"id":99,"title":"broken","state":3,"percent":0,"num_done":0,"num_err":0}`,
	}
	client := newTestClient(t, fixtures)

	_, err := client.PollTask(t.Context(), 99)
	if err == nil {
		t.Fatal("expected error for state=3 (failed)")
	}
}

func TestWaitForStableResults_AlreadyStable(t *testing.T) {
	hits := `{"firewall policy":[{"policyid":1,"name":"r1","hitcount":5}]}`
	client := newTestClientFull(t, nil, nil, map[string]func(map[string]any) string{
		"sys/task/result": func(map[string]any) string { return hits },
	})

	initial := []PolicyHitCount{{PolicyID: 1, Name: "r1", HitCount: 5}}
	result, err := client.waitForStableResults(t.Context(), 1, initial)
	if err != nil {
		t.Fatalf("waitForStableResults: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}
}

func TestGetHitCountResults_ParsesEnvelope(t *testing.T) {
	hits := `{"firewall policy":[{"policyid":10,"name":"web","hitcount":999,"first_hit":1000,"last_hit":2000,"uuid":"abc-123"}]}`
	client := newTestClientFull(t, nil, nil, map[string]func(map[string]any) string{
		"sys/task/result": func(map[string]any) string { return hits },
	})

	results, err := client.GetHitCountResults(t.Context(), 1)
	if err != nil {
		t.Fatalf("GetHitCountResults: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	p := results[0]
	if p.PolicyID != 10 || p.HitCount != 999 || p.UUID != "abc-123" {
		b, _ := json.Marshal(p)
		t.Fatalf("unexpected policy: %s", b)
	}
}
