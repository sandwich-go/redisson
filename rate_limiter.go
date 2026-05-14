package redisson

import (
	"context"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislimiter"
)

type (
	RateLimiterResult = rueidislimiter.Result
	RateLimitOption   = rueidislimiter.RateLimitOption
)

var WithRateLimitOption = func(limit int, window time.Duration) RateLimitOption {
	return rueidislimiter.WithCustomRateLimit(limit, window)
}

type RateLimiter interface {
	// Check if a request is allowed under the rate limit without incrementing the count.
	Check(ctx context.Context, identifier string, options ...RateLimitOption) (RateLimiterResult, error)
	// Allow a single request, incrementing the counter if allowed.
	Allow(ctx context.Context, identifier string, options ...RateLimitOption) (RateLimiterResult, error)
	// AllowN n requests, incrementing the counter accordingly if allowed.
	AllowN(ctx context.Context, identifier string, n int64, options ...RateLimitOption) (RateLimiterResult, error)
	// Limit return maximum number of allowed requests per window.
	Limit() int
}

type wrapRateLimiter struct {
	v ConfInterface
	rueidislimiter.RateLimiterClient
}

func (w *wrapRateLimiter) Limit() int { return w.RateLimiterClient.Limit() }
func (w *wrapRateLimiter) Check(ctx context.Context, identifier string, options ...RateLimitOption) (RateLimiterResult, error) {
	return w.RateLimiterClient.Check(ctx, identifier, options...)
}

func (w *wrapRateLimiter) Allow(ctx context.Context, identifier string, options ...RateLimitOption) (RateLimiterResult, error) {
	return w.RateLimiterClient.Allow(ctx, identifier, options...)
}

func (w *wrapRateLimiter) AllowN(ctx context.Context, identifier string, n int64, options ...RateLimitOption) (RateLimiterResult, error) {
	return w.RateLimiterClient.AllowN(ctx, identifier, n, options...)
}

func newRateLimiter(c *client, opts ...RateLimiterOption) (RateLimiter, error) {
	cc := newRateLimiterOptions(opts...)
	l, err := rueidislimiter.NewRateLimiter(rueidislimiter.RateLimiterOption{
		ClientBuilder: func(option rueidis.ClientOption) (rueidis.Client, error) {
			return rueidis.NewClient(option)
		},
		ClientOption: confVisitor2ClientOption(c.v),
		KeyPrefix:    cc.GetKeyPrefix(),
		Limit:        cc.GetLimit(),
		Window:       cc.GetWindow(),
	})
	if err != nil {
		return nil, err
	}
	return &wrapRateLimiter{RateLimiterClient: l, v: c.v}, nil
}

func (c *client) NewRateLimiter(opts ...RateLimiterOption) (RateLimiter, error) {
	return newRateLimiter(c, opts...)
}
