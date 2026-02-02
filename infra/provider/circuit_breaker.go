package provider

import (
	"context"
	"errors"
	"sync"
	"time"

	"search-engine/domain"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	threshold       int
	timeout         time.Duration
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	state := cb.state
	lastFailure := cb.lastFailureTime
	cb.mu.RUnlock()

	switch state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(lastFailure) > cb.timeout {
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		return
	}

	if cb.failureCount >= cb.threshold {
		cb.state = StateOpen
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount > 0 {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

type CircuitBreakerProvider struct {
	provider ContentProvider
	breaker  *CircuitBreaker
}

func NewCircuitBreakerProvider(provider ContentProvider, threshold int, timeout time.Duration) *CircuitBreakerProvider {
	return &CircuitBreakerProvider{
		provider: provider,
		breaker:  NewCircuitBreaker(threshold, timeout),
	}
}

func (p *CircuitBreakerProvider) Search(ctx context.Context, query string) ([]domain.ProviderContent, error) {
	var contents []domain.ProviderContent
	var searchErr error

	err := p.breaker.Execute(func() error {
		var err error
		contents, err = p.provider.Search(ctx, query)
		searchErr = err
		return err
	})

	if errors.Is(err, ErrCircuitOpen) {
		return nil, err
	}

	return contents, searchErr
}

func (p *CircuitBreakerProvider) Name() string {
	return p.provider.Name()
}

func (p *CircuitBreakerProvider) HealthCheck(ctx context.Context) error {
	if p.breaker.State() == StateOpen {
		return ErrCircuitOpen
	}
	return p.provider.HealthCheck(ctx)
}

func (p *CircuitBreakerProvider) CircuitState() CircuitState {
	return p.breaker.State()
}
