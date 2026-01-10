// Package core provides core infrastructure for TGCP including authentication,
// HTTP client configuration, caching, service registry, and project management.
//
// HTTP Client Configuration
//
// The HTTP client is configured with three layers of middleware:
//  1. Google Authentication (Application Default Credentials)
//  2. Rate Limiting (Token Bucket Algorithm)
//  3. Retry Logic (Exponential Backoff)
//
// Rate Limiting
//
// The rate limiter uses a token bucket algorithm to prevent API quota exhaustion.
// Configuration: 10 requests per second with a burst capacity of 20 requests.
//
// How it works:
//   - Tokens are added to the bucket at a constant rate (10 tokens/second)
//   - Each request consumes 1 token
//   - If tokens are available, the request proceeds immediately
//   - If no tokens are available, the request waits until a token is available
//   - Burst capacity allows handling traffic spikes up to 20 requests
//
// Example:
//   client, err := core.NewHTTPClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
//   if err != nil {
//       return err
//   }
//   // All requests through this client are automatically rate-limited
//
// Retry Logic
//
// The retry transport implements exponential backoff for transient failures.
// Configuration: Maximum 3 retries with exponential backoff (100ms, 200ms, 400ms).
//
// Retry conditions:
//   - Network errors (connection failures, timeouts)
//   - HTTP 429 (Too Many Requests) - rate limit exceeded
//   - HTTP 5xx (Server Errors) - temporary server issues
//
// Non-retryable conditions:
//   - HTTP 4xx (Client Errors) - except 429
//   - Context cancellation
//   - Maximum retries exceeded
//
// Backoff calculation: 2^i * 100ms where i is the retry attempt (0, 1, 2)
//   - Attempt 1: 100ms delay
//   - Attempt 2: 200ms delay
//   - Attempt 3: 400ms delay
//
// Example:
//   // A request that fails with 500 will be retried up to 3 times
//   // with increasing delays between attempts
//   resp, err := client.Get("https://compute.googleapis.com/...")
//
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
//
// The client uses a middleware chain: Client -> RateLimit -> Retry -> Auth(Base)
// All requests through this client are automatically rate-limited and retried on transient failures.
//
// Example:
//   ctx := context.Background()
//   client, err := core.NewHTTPClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
//   if err != nil {
//       return fmt.Errorf("failed to create HTTP client: %w", err)
//   }
//   // Use client for GCP API calls - rate limiting and retries are automatic
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

// TokenBucket implements a token bucket rate limiting algorithm.
// It allows a certain number of requests per second (rate) with a burst capacity.
//
// Algorithm:
//   - Tokens are added to the bucket at a constant rate
//   - Each request consumes one token
//   - If tokens are available, the request proceeds immediately
//   - If no tokens are available, the request waits until tokens are refilled
//
// Thread-safe: Uses mutex to protect concurrent access to token state.
type TokenBucket struct {
	rate       float64 // tokens per second
	burst      float64 // maximum tokens (burst capacity)
	tokens     float64 // current number of tokens
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter.
//
// Parameters:
//   - rate: Tokens added per second (e.g., 10.0 = 10 requests/second)
//   - burst: Maximum tokens (e.g., 20.0 = can handle 20 requests in quick succession)
//
// Example:
//   limiter := NewTokenBucket(10.0, 20.0) // 10 req/s, burst of 20
func NewTokenBucket(rate, burst float64) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     burst,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled.
// It automatically refills tokens based on elapsed time since last refill.
//
// Returns:
//   - nil if a token was successfully acquired
//   - context.Err() if the context was cancelled
//
// Example:
//   if err := limiter.Wait(ctx); err != nil {
//       return err // Context cancelled
//   }
//   // Token acquired, proceed with request
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

// RetryTransport implements automatic retry with exponential backoff for HTTP requests.
// It wraps an existing http.RoundTripper and retries failed requests up to 3 times.
//
// Retry conditions:
//   - Network errors (connection failures, timeouts)
//   - HTTP 429 (Too Many Requests)
//   - HTTP 5xx (Server Errors)
//
// Non-retryable:
//   - HTTP 4xx (Client Errors) except 429
//   - Context cancellation
//   - Maximum retries exceeded
//
// Backoff: Exponential backoff with formula 2^i * 100ms
//   - Retry 1: 100ms delay
//   - Retry 2: 200ms delay
//   - Retry 3: 400ms delay
type RetryTransport struct {
	Next http.RoundTripper
}

// RoundTrip executes the HTTP request with automatic retry on transient failures.
// It implements exponential backoff between retry attempts.
//
// The method will retry up to 3 times for:
//   - Network errors
//   - HTTP 429 (rate limit)
//   - HTTP 5xx (server errors)
//
// Returns the last response/error if all retries are exhausted.
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
