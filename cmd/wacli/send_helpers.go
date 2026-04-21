package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/steipete/wacli/internal/app"
	"go.mau.fi/whatsmeow"
)

const sendAttemptTimeout = 45 * time.Second

func runSendOperation[T any](
	ctx context.Context,
	reconnect func(context.Context) error,
	op func(context.Context) (T, error),
) (T, error) {
	result, err := runSendAttempt(ctx, sendAttemptTimeout, op)
	if err == nil {
		return result, nil
	}

	var zero T
	if !isRetryableSendError(err) || ctx.Err() != nil {
		return zero, err
	}
	if reconnectErr := reconnect(ctx); reconnectErr != nil {
		return zero, fmt.Errorf("%w; reconnect failed: %v", err, reconnectErr)
	}
	return runSendAttempt(ctx, sendAttemptTimeout, op)
}

func runSendAttempt[T any](ctx context.Context, timeout time.Duration, op func(context.Context) (T, error)) (T, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		value T
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		value, err := op(attemptCtx)
		ch <- result{value: value, err: err}
	}()

	select {
	case res := <-ch:
		if errors.Is(res.err, context.DeadlineExceeded) && errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			var zero T
			return zero, fmt.Errorf("send timed out after %s", timeout)
		}
		return res.value, res.err
	case <-attemptCtx.Done():
		var zero T
		if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			return zero, fmt.Errorf("send timed out after %s", timeout)
		}
		return zero, attemptCtx.Err()
	}
}

func isRetryableSendError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, whatsmeow.ErrIQTimedOut) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "failed to send usync query") ||
		strings.Contains(msg, "failed to get user info") ||
		strings.Contains(msg, "failed to get device list") ||
		strings.Contains(msg, "info query timed out") ||
		strings.Contains(msg, "not connected")
}

func reconnectForSend(a interface {
	WA() app.WAClient
	Connect(context.Context, bool, func(string)) error
}) func(context.Context) error {
	return func(ctx context.Context) error {
		a.WA().Close()
		return a.Connect(ctx, false, nil)
	}
}
