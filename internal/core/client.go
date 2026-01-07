package core

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
)

// NewHTTPClient returns an http.Client configured with:
// 1. Google Authentication (ADC)
// 2. Rate Limiting (10 RPS, Burst 20)
// 3. Retry Logic (Exponential Backoff, 3 retries)
func NewHTTPClient(ctx context.Context, scopes ...string) (*http.Client, error) {
	// 1. Create the base authenticated client
	client, err := google.DefaultClient(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create default client: %w", err)
	}

	// 2. Wrap the transport
	// The chain will be: Client -> RateLimit -> Retry -> Auth(Base)
	// We wrap in reverse order because we wrap the *existing* transport.
	// So: NewTransport calls ExistingTransport.
	// Existing is Auth+Base.
	// Retry wraps Existing.
	// RateLimit wraps Retry.

	baseTransport := client.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	retryTransport := &RetryTransport{
		Next: baseTransport,
	}

	rateLimitTransport := &RateLimitTransport{
		Next:    retryTransport,
		Limiter: NewTokenBucket(10, 20), // 10 req/s, 20 burst
	}

	client.Transport = rateLimitTransport
	return client, nil
}

// --- Rate Limiter ---

type TokenBucket struct {
	rate       float64 // tokens per second
	burst      float64
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(rate, burst float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     burst,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for {
		now := time.Now()
		// Refill
		elapsed := now.Sub(tb.lastRefill).Seconds()
		tb.tokens = math.Min(tb.burst, tb.tokens+(elapsed*tb.rate))
		tb.lastRefill = now

		if tb.tokens >= 1.0 {
			tb.tokens -= 1.0
			return nil
		}

		// Wait for enough tokens
		missing := 1.0 - tb.tokens
		waitTime := time.Duration((missing / tb.rate) * float64(time.Second))

		// Unlock to wait
		tb.mu.Unlock()
		select {
		case <-ctx.Done():
			tb.mu.Lock() // Re-lock just to defer unlock safely, though we return error
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop to re-check/claim
		}
		tb.mu.Lock()
	}
}

type RateLimitTransport struct {
	Next    http.RoundTripper
	Limiter *TokenBucket
}

func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.Limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.Next.RoundTrip(req)
}

// --- Retry Transport ---

type RetryTransport struct {
	Next http.RoundTripper
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxRetries := 3
	var resp *http.Response
	var err error

	for i := 0; i <= maxRetries; i++ {
		resp, err = t.Next.RoundTrip(req)

		// Success case
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// Check if we should retry
		shouldRetry := false
		if err != nil {
			// Network errors are usually retryable
			shouldRetry = true
		} else if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			shouldRetry = true
			resp.Body.Close() // Close body before retrying
		}

		if !shouldRetry || i == maxRetries {
			break
		}

		// Calculate backoff: 2^i * 100ms
		backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
		// Jitter could be added here

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
			// Continue and retry
		}
	}

	return resp, err
}
