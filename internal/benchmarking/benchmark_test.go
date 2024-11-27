package benchmarking

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunBenchmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))

	}))
	defer server.Close()

	config := BecnhmarkConfig{
		RequestCount:     5,
		MaxConcurrency:   2,
		RequestTimeout:   1 * time.Second,
		BenchmarkTimeout: 30 * time.Second,
		URL:              server.URL,
	}
	client := &http.Client{}

	result, err := RunBenchmark(client, config)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result.Results))
	assert.True(t, result.Duration > 0)

	for _, res := range result.Results {
		assert.NoError(t, res.Error)
		assert.True(t, res.Duration > 0)
		assert.True(t, res.Bytes > 0)
	}
}

func TestRunBenchmark_RequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := BecnhmarkConfig{
		RequestCount:     3,
		MaxConcurrency:   2,
		RequestTimeout:   500 * time.Millisecond,
		BenchmarkTimeout: 30 * time.Second,
		URL:              server.URL,
	}
	client := &http.Client{}

	result, err := RunBenchmark(client, config)
	assert.ErrorIs(t, err, ErrRequestTimeoutExceeded)
	assert.Equal(t, 0, len(result.Results))
}

func TestRunBenchmark_BenchmarkTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := BecnhmarkConfig{
		RequestCount:     3,
		MaxConcurrency:   2,
		RequestTimeout:   3 * time.Second,
		BenchmarkTimeout: 1 * time.Second,
		URL:              server.URL,
	}
	client := &http.Client{}

	result, err := RunBenchmark(client, config)
	assert.ErrorIs(t, err, ErrBenchmarkTimeoutExcceded)
	assert.Equal(t, 0, len(result.Results))
}

