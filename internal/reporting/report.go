package reporting

import (
	"fmt"
	"time"

	"github.com/AntonGavrilov/spikeys/internal/benchmarking"
)

type Result struct {
	Duration time.Duration
	Bytes    int64
	Error    error
}

type Summary struct {
	CompletedRequests  int64
	FailedRequests     int64
	TotalBytes         int64
	MeanTimePerRequest time.Duration
	RequestsPerSecond  float64
	TotalDuration      time.Duration
	ConcurrencyLevel   int64
}

func GenerateSummary(results []*benchmarking.RequestResult, totalDuration time.Duration, concurrency int64) Summary {
	var completedRequests, failedRequests int64
	var totalBytes int64
	var totalRequestTime time.Duration

	for _, res := range results {
		if res != nil {
			if res.Error == nil {
				completedRequests++
				totalBytes += res.Bytes
				totalRequestTime += res.Duration
			} else {
				failedRequests++
			}
		}
	}

	meanTimePerRequest := time.Duration(0)
	if completedRequests > 0 {
		meanTimePerRequest = totalRequestTime / time.Duration(completedRequests)
	}

	requestsPerSecond := 0.0
	if totalDuration > 0 {
		requestsPerSecond = float64(completedRequests) / totalDuration.Seconds()
	}

	return Summary{
		CompletedRequests:  completedRequests,
		FailedRequests:     failedRequests,
		TotalBytes:         totalBytes,
		MeanTimePerRequest: meanTimePerRequest,
		RequestsPerSecond:  requestsPerSecond,
		TotalDuration:      totalDuration,
		ConcurrencyLevel:   concurrency,
	}
}

func PrintSummary(summary Summary) {
	fmt.Println()
	fmt.Println("Benchmark Results")
	fmt.Println("-----------------")
	printItem("Completed requests:", summary.CompletedRequests)
	printItem("Failed requests:", summary.FailedRequests)
	printItem("Total transferred:", fmt.Sprintf("%d bytes", summary.TotalBytes))
	printItem("Concurrency level:", summary.ConcurrencyLevel)
	printItem("Time taken for tests:", fmt.Sprintf("%.3f seconds", summary.TotalDuration.Seconds()))
	printItem("Time per request (mean):", fmt.Sprintf("%.3f ms", float64(summary.MeanTimePerRequest.Microseconds())/1000))
	printItem("Requests per second:", fmt.Sprintf("%.3f [#/sec]", summary.RequestsPerSecond))
	fmt.Println()
}

func printItem(key string, value interface{}) {
	const padding = 25
	spaces := padding - len(key)
	if spaces < 1 {
		spaces = 1
	}
	fmt.Printf("%s%s %v\n", key, spacesStr(spaces), value)
}

func spacesStr(n int) string {
	return fmt.Sprintf("%*s", n, "")
}
