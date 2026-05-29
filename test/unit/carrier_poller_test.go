package unit

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/dinocodesx/subscription-reconciler/internal/domain/carrier"
)

type mockPollRepo struct {
	claimedUsers []string
	claimErr     error
	appliedCalls []applyCall
	applyErr     error
}

type applyCall struct {
	userID string
	status string
}

func (m *mockPollRepo) ClaimCarrierUsers(_ context.Context, _ int) ([]string, error) {
	return m.claimedUsers, m.claimErr
}

func (m *mockPollRepo) ApplyCarrierStatus(_ context.Context, userID, status string) error {
	m.appliedCalls = append(m.appliedCalls, applyCall{userID, status})
	return m.applyErr
}

type mockCarrierClient struct {
	responses map[string]string
	errs      map[string]error
}

func (m *mockCarrierClient) PlanStatus(_ context.Context, userID string) (string, error) {
	if err, ok := m.errs[userID]; ok {
		return "", err
	}
	return m.responses[userID], nil
}

func TestPollerActiveStatusApplied(t *testing.T) {
	repo := &mockPollRepo{claimedUsers: []string{"u_1"}}
	client := &mockCarrierClient{
		responses: map[string]string{"u_1": "active"},
		errs:      map[string]error{},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.appliedCalls) != 1 {
		t.Fatalf("expected 1 apply call, got %d", len(repo.appliedCalls))
	}
	if repo.appliedCalls[0].status != "active" {
		t.Fatalf("expected status 'active', got %q", repo.appliedCalls[0].status)
	}
}

func TestPollerInactiveStatusApplied(t *testing.T) {
	repo := &mockPollRepo{claimedUsers: []string{"u_1"}}
	client := &mockCarrierClient{
		responses: map[string]string{"u_1": "inactive"},
		errs:      map[string]error{},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.appliedCalls) != 1 {
		t.Fatalf("expected 1 apply call, got %d", len(repo.appliedCalls))
	}
	if repo.appliedCalls[0].status != "inactive" {
		t.Fatalf("expected status 'inactive', got %q", repo.appliedCalls[0].status)
	}
}

func TestPollerAPIErrorPassedToRepoApply(t *testing.T) {
	repo := &mockPollRepo{claimedUsers: []string{"u_1"}}
	client := &mockCarrierClient{
		responses: map[string]string{"u_1": "api_error"},
		errs:      map[string]error{},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// api_error is passed through to the repo's ApplyCarrierStatus;
	// the *repo* implementation decides to skip on api_error. The poller itself
	// still calls Apply — the skip logic is in the repository layer.
	if len(repo.appliedCalls) != 1 {
		t.Fatalf("expected 1 apply call (repo handles api_error), got %d", len(repo.appliedCalls))
	}
}

func TestPollerClientErrorContinuesRemaining(t *testing.T) {
	repo := &mockPollRepo{claimedUsers: []string{"u_fail", "u_ok"}}
	client := &mockCarrierClient{
		responses: map[string]string{"u_ok": "active"},
		errs:      map[string]error{"u_fail": errors.New("connection refused")},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// u_fail should have been skipped, u_ok should have been applied.
	if len(repo.appliedCalls) != 1 {
		t.Fatalf("expected 1 apply call, got %d", len(repo.appliedCalls))
	}
	if repo.appliedCalls[0].userID != "u_ok" {
		t.Fatalf("expected u_ok applied, got %q", repo.appliedCalls[0].userID)
	}
}

func TestPollerClaimErrorPropagates(t *testing.T) {
	repo := &mockPollRepo{claimErr: errors.New("db down")}
	client := &mockCarrierClient{
		responses: map[string]string{},
		errs:      map[string]error{},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	err := poller.RunOnce(context.Background())
	if err == nil {
		t.Fatalf("expected error from claim failure")
	}
}

func TestPollerNoClaims(t *testing.T) {
	repo := &mockPollRepo{claimedUsers: []string{}}
	client := &mockCarrierClient{
		responses: map[string]string{},
		errs:      map[string]error{},
	}

	poller := carrier.NewPoller(repo, client, 10, log.Default())
	if err := poller.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.appliedCalls) != 0 {
		t.Fatalf("expected no apply calls, got %d", len(repo.appliedCalls))
	}
}
