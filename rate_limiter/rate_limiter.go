package rate_limiter

import "time"

type RateLimiter struct {
	tick   *time.Ticker
	bucket chan struct{}
	cancel chan struct{}
	buffer int
}

func NewRateLimiter(qps int, maxPeekBuffer int) RateLimiter {
	r := RateLimiter{
		tick:   time.NewTicker(time.Second / time.Duration(qps)),
		bucket: make(chan struct{}, maxPeekBuffer),
		cancel: make(chan struct{}),
		buffer: maxPeekBuffer,
	}
	r.start()
	return r
}

func (r *RateLimiter) start() {
	for i := 0; i < r.buffer; i++ {
		r.bucket <- struct{}{}
	}
	go func() {
		for range r.tick.C {
			select {
			case r.bucket <- struct{}{}:
			case <-r.cancel:
				r.tick.Stop()
				return
			}
		}
	}()
}

func (r *RateLimiter) Stop() {
	close(r.cancel)
}

func (r *RateLimiter) Call(call func() error) error {
	<-r.bucket
	return call()
}
