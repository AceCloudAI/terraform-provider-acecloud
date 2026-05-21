// Package wait provides reusable polling and retry helpers for Terraform
// provider resources that need to wait for async operations or retry on
// transient errors.
//
// Three core patterns are implemented:
//
//   - WaitForStatus: poll a resource by ID until it reaches a target status
//   - RetryOnConflict: retry a mutating operation when the resource is in
//     an immutable/pending state
//   - PollForResource: poll a list endpoint until a newly-created async
//     resource appears (matched by a caller-supplied predicate)
package wait

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Default timeouts and intervals.
const (
	DefaultStatusTimeout  = 5 * time.Minute  // WaitForStatus
	DefaultStatusInterval = 5 * time.Second  // WaitForStatus poll interval
	DefaultRetryTimeout   = 2 * time.Minute  // RetryOnConflict — 8 attempts × 15s budget
	DefaultRetryInterval  = 15 * time.Second // RetryOnConflict retry interval (post-backoff)
	DefaultPollTimeout    = 2 * time.Minute  // PollForResource
	DefaultPollInterval   = 5 * time.Second  // PollForResource poll interval
)

// StatusResult is returned by the refresh function passed to WaitForStatus.
type StatusResult struct {
	Status string      // Current provisioning status (e.g. "ACTIVE", "BUILD")
	Data   interface{} // Arbitrary payload returned to the caller
}

// RefreshFunc fetches the current status of a resource. It should return the
// current StatusResult, or an error if the fetch failed. Transient fetch
// errors (e.g. network blip) are tolerated — the poller will continue.
type RefreshFunc func(ctx context.Context) (*StatusResult, error)

// WaitForStatusOpts configures WaitForStatus.
type WaitForStatusOpts struct {
	Refresh      RefreshFunc   // Required: function that returns current status
	TargetStatus []string      // Statuses that mean "done" (e.g. ["ACTIVE"])
	ErrorStatus  []string      // Statuses that mean "failed" (e.g. ["ERROR"])
	Timeout      time.Duration // Overall timeout (0 = DefaultStatusTimeout)
	PollInterval time.Duration // Interval between polls (0 = DefaultStatusInterval)
}

// WaitForStatus polls a resource until it reaches one of the target or error
// statuses, or until the timeout expires.
//
// Returns the final StatusResult and nil on success, or nil and an error on
// timeout/error status/context cancellation.
func WaitForStatus(ctx context.Context, opts WaitForStatusOpts) (*StatusResult, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultStatusTimeout
	}
	interval := opts.PollInterval
	if interval == 0 {
		interval = DefaultStatusInterval
	}

	targetSet := toSet(opts.TargetStatus)
	errorSet := toSet(opts.ErrorStatus)

	deadline := time.After(timeout)
	for {
		result, err := opts.Refresh(ctx)
		if err == nil && result != nil {
			if _, ok := targetSet[result.Status]; ok {
				return result, nil
			}
			if _, ok := errorSet[result.Status]; ok {
				return result, fmt.Errorf("resource entered error status: %s", result.Status)
			}
		}
		// err != nil is treated as transient — keep polling

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-deadline:
			status := "unknown"
			if result != nil {
				status = result.Status
			}
			return result, fmt.Errorf("timed out waiting for target status %v (last status: %s)", opts.TargetStatus, status)
		case <-time.After(interval):
			// next iteration
		}
	}
}

// RetryOnConflict retries a mutating operation (create, delete) when the
// resource or parent resource is in an immutable/pending state. The operation
// function should return nil on success.
//
// Retryable errors are detected by checking if the error message contains any
// of the default transient substrings ("immutable", "PENDING_", "Please try
// again", "in use", "SecurityGroupInUse"), or custom substrings passed via
// RetryableErrors.
//
// When RetryAuthErrors is set (intended for Delete paths), the helper also
// retries transient auth-related errors ("401", "Unauthorized", "Unable to
// process this request") that can surface during concurrent destroys.
//
// The first 3 attempts are quick probes (1s, 3s, 7s) that recover fast from
// brief transient blips. Subsequent attempts use the configured interval
// (default 15s) so we don't flood the platform if the issue persists.
func RetryOnConflict(ctx context.Context, opts RetryOnConflictOpts) error {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultRetryTimeout
	}
	baseInterval := opts.Interval
	if baseInterval == 0 {
		baseInterval = DefaultRetryInterval
	}

	retryable := opts.RetryableErrors
	if len(retryable) == 0 {
		retryable = defaultRetryableErrors
	}
	if opts.RetryAuthErrors {
		retryable = append(retryable, authRetryableErrors...)
	}

	deadline := time.After(timeout)
	attempt := 0
	for {
		err := opts.Operation(ctx)
		if err == nil {
			return nil
		}

		if !isRetryable(err, retryable) {
			return err
		}

		attempt++
		wait := backoffInterval(attempt, baseInterval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timed out retrying operation: %w", err)
		case <-time.After(wait):
			// next attempt
		}
	}
}

// RetryOnConflictOpts configures RetryOnConflict.
type RetryOnConflictOpts struct {
	Operation       func(ctx context.Context) error // Required: the operation to retry
	Timeout         time.Duration                   // Overall timeout (0 = DefaultRetryTimeout)
	Interval        time.Duration                   // Interval between retries (0 = DefaultRetryInterval); first 3 attempts use exponential backoff
	RetryableErrors []string                        // Error substrings that trigger retry (nil = defaults)
	RetryAuthErrors bool                            // Also retry 401/Unauthorized/"Unable to process this request" — intended for Delete paths during concurrent destroys
}

var defaultRetryableErrors = []string{
	"immutable",
	"PENDING_",
	"Please try again",
	"in use",
	"SecurityGroupInUse",
}

// authRetryableErrors are transient auth-shaped errors that can surface
// during concurrent destroys. Opt-in via RetryOnConflictOpts.RetryAuthErrors
// (Delete paths only — Create/Update must surface real auth errors
// immediately so genuinely-bad credentials don't disappear into a retry).
var authRetryableErrors = []string{
	"401",
	"Unauthorized",
	"Unable to process this request",
}

// backoffInterval returns the wait duration before retry attempt n.
//
// Attempts 1-3 are quick probes scaled relative to the base interval
// (base/15, base/5, ~base/2 ≈ 1s/3s/7s when base=15s) so we recover fast
// from transient blips that clear in under a second. Attempts 4+ use the
// configured base interval so we don't flood the platform if the issue
// persists. With base=15s the full retry budget is roughly:
//   attempt 1 (after 1s) + attempt 2 (after 3s) + attempt 3 (after 7s)
//   + attempts 4-8 (after 15s each) ≈ 86s
// well inside the 2-minute default deadline.
func backoffInterval(attempt int, base time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	switch attempt {
	case 1:
		return base / 15
	case 2:
		return base / 5
	case 3:
		return base / 2
	default:
		return base
	}
}

// PollForResourceOpts configures PollForResource.
type PollForResourceOpts struct {
	List         func(ctx context.Context) (interface{}, error) // Required: fetches list and returns matched item or nil
	Timeout      time.Duration                                  // Overall timeout (0 = DefaultPollTimeout)
	PollInterval time.Duration                                  // Interval between polls (0 = DefaultPollInterval)
}

// PollForResource polls a list endpoint until a newly-created resource appears.
// The List function should return the matched item (non-nil) when found, or nil
// if not yet present. Returns the matched item or an error on timeout.
func PollForResource(ctx context.Context, opts PollForResourceOpts) (interface{}, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultPollTimeout
	}
	interval := opts.PollInterval
	if interval == 0 {
		interval = DefaultPollInterval
	}

	deadline := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("timed out waiting for resource to appear")
		case <-time.After(interval):
			// poll
		}

		item, err := opts.List(ctx)
		if err != nil {
			continue // transient error, keep polling
		}
		if item != nil {
			return item, nil
		}
	}
}

func toSet(items []string) map[string]struct{} {
	s := make(map[string]struct{}, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func isRetryable(err error, retryable []string) bool {
	msg := err.Error()
	for _, substr := range retryable {
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}
