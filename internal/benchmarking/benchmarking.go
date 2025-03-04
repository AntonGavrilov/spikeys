package benchmarking

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type RequestResult struct {
	Duration time.Duration
	Bytes    int64
	Error    error
}
type Result struct {
	Results  []*RequestResult
	Duration time.Duration
}

type DefaultBenchmarker struct {
	performer RequestPerformer
}

func NewBenchmarker(preformer RequestPerformer) DefaultBenchmarker {
	return DefaultBenchmarker{preformer}
}

type RequestPerformer interface {
	preformRequest(ctx context.Context, address string) RequestResult
}
type DefaultRequestPerformer struct {
	HTTPClient *http.Client
}

func (b *DefaultBenchmarker) Run(config BenchmarkConfig) (Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
	defer cancel()

	sem := semaphore.NewWeighted(int64(config.MaxConcurrency))
	results := make([]*RequestResult, config.RequestCount)

	var startTime time.Time
	doneChan := make(chan struct{})
	reqTimeoutChan := make(chan struct{}, config.MaxConcurrency)

	go func() {
		defer close(reqTimeoutChan)
		wg := sync.WaitGroup{}
		startTime = time.Now()
		for i := 0; i < config.RequestCount; i++ {
			err := sem.Acquire(ctx, 1)
			if err != nil {
				break
			}
			wg.Add(1)
			go func(num int) {
				defer func() {
					sem.Release(1)
					wg.Done()
				}()
				ctx, cancelReq := context.WithTimeout(ctx, config.RequestTimeout)
				defer cancelReq()
				result := b.performer.preformRequest(ctx, config.URL)
				if result.Error != nil && errors.Is(result.Error, context.DeadlineExceeded) {
					reqTimeoutChan <- struct{}{}
					cancel()
					return
				}
				results[num] = &result
			}(i)
		}
		wg.Wait()
		close(doneChan)
	}()
	select {
	case <-reqTimeoutChan:
		return Result{}, ErrRequestTimeoutExceeded
	case <-ctx.Done():
		return Result{}, ErrBenchmarkTimeoutExcceded
	case <-doneChan:
		return Result{results, time.Since(startTime)}, nil
	}
}

func (performer DefaultRequestPerformer) preformRequest(ctx context.Context, address string) RequestResult {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)

	if err != nil {
		return RequestResult{time.Since(start), 0, err}
	}
	resp, err := performer.HTTPClient.Do(req)
	if err != nil {
		return RequestResult{time.Since(start), 0, err}
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return RequestResult{time.Since(start), 0, errors.New("not ok error")}
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return RequestResult{time.Since(start), 0, err}
	}
	return RequestResult{time.Since(start), int64(len(content)), nil}
}
