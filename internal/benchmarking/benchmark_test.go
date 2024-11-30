package benchmarking

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRunBenchmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))

	}))
	defer server.Close()
	config := BenchmarkConfig{
		RequestCount:     5,
		MaxConcurrency:   2,
		RequestTimeout:   1 * time.Second,
		BenchmarkTimeout: 30 * time.Second,
		URL:              server.URL,
	}
	p := DefaultRequestPerformer{&http.Client{}}
	benchmarker := NewBenchmarker(&p)
	result, err := benchmarker.Run(config)

	require.NoError(t, err)
	assert.Len(t, result.Results, 5 )
	assert.Positive(t, result.Duration)

	for _, res := range result.Results {
		require.NoError(t, res.Error)
		assert.Positive(t, res.Duration)
		assert.Positive(t, res.Bytes)
	}
}

func TestRunBenchmark_RequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := BenchmarkConfig{
		RequestCount:     3,
		MaxConcurrency:   2,
		RequestTimeout:   500 * time.Millisecond,
		BenchmarkTimeout: 30 * time.Second,
		URL:              server.URL,
	}
	p := DefaultRequestPerformer{&http.Client{}}
	benchmarker := NewBenchmarker(&p)
	result, err := benchmarker.Run(config)
	require.ErrorIs(t, err, ErrRequestTimeoutExceeded)
	assert.Empty(t, result.Results)
}

func TestRunBenchmark_BenchmarkTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := BenchmarkConfig{
		RequestCount:     3,
		MaxConcurrency:   2,
		RequestTimeout:   3 * time.Second,
		BenchmarkTimeout: 1 * time.Second,
		URL:              server.URL,
	}
	p := DefaultRequestPerformer{&http.Client{}}

	benchmarker := NewBenchmarker(&p)
	result, err := benchmarker.Run(config)
	require.ErrorIs(t, err, ErrBenchmarkTimeoutExcceded)
	assert.Empty(t, result.Results)
}

func TestRunBenchmark_MaxConcurrency(t *testing.T) {
	const maxConcurrency = 5
	const requestCount = 20
	var maxActiveGoroutines int64
	var currentActiveGoroutines int64
	var mu sync.Mutex

	config := BenchmarkConfig{
		BenchmarkTimeout: 5 * time.Second,
		RequestTimeout:   1 * time.Second,
		MaxConcurrency:   maxConcurrency,
		RequestCount:     requestCount,
		URL:              "http://example.com",
	}
	m := RequestResult{
		Duration: 50 * time.Millisecond,
		Bytes:    100,
		Error:    nil,
	}
	p := new(RequestPerformerMock)
	p.On("preformRequest", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		mu.Lock()
		currentActiveGoroutines++
		if currentActiveGoroutines > maxActiveGoroutines {
			maxActiveGoroutines = currentActiveGoroutines
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		currentActiveGoroutines--
		mu.Unlock()
	}).Return(m)
	b := NewBenchmarker(p)
	_, err := b.Run(config)

	require.NoError(t, err, "Benchmark execution should not fail")

	assert.LessOrEqual(t, maxActiveGoroutines, int64(maxConcurrency),
		"Max active goroutines exceeded: got %d, want <= %d", maxActiveGoroutines, maxConcurrency)
}

func TestRunBenchmark_RequestsWithErrorsShouldContainsInResult(t *testing.T) {
	const requestCount = 5
	config := BenchmarkConfig{
		BenchmarkTimeout: 30 * time.Second,
		RequestTimeout:   30 * time.Second,
		MaxConcurrency:   5,
		RequestCount:     5,
		URL:              "http://example.com",
	}
	p := new(RequestPerformerMock)
	rs := RequestResult{Error: errors.New("error")}
	p.On("preformRequest", mock.Anything, mock.Anything).Return(rs)
	b := NewBenchmarker(p)

	result, err := b.Run(config)
	require.NoError(t, err, "Benchmark execution should not fail")
	assert.Len(t, result.Results, requestCount, "Results should contain exactly %d entries", requestCount)
	for i, res := range result.Results {
		require.NotNil(t, res, "Result %d is nil", i)
		require.Error(t, res.Error, "Result %d should have an error", i)
	}
	p.AssertNumberOfCalls(t, "preformRequest", requestCount)
}

type RequestPerformerMock struct {
	mock.Mock
}

func (m *RequestPerformerMock) preformRequest(ctx context.Context, address string) RequestResult {
	args := m.Called(ctx, address)
	return args.Get(0).(RequestResult)
}
