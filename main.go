package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"
)

type result struct {
	Time  time.Duration
	Bytes int64
	err   error
}

var reqCount int = 1
var ComplitedRequest = 0
var maxConcurrency int64 = 1
var benchmarkTimelimit int = 30 //seconds
var reqTimeout int = 30         //seconds
var sem *semaphore.Weighted
var results []*result
var testDuration time.Duration
var TotalTransferredBytes int64
var httpClient *http.Client = &http.Client{}

func flagSet(f string) bool {
	return viper.IsSet(f)
}
func main() {

	pflag.IntP("requests", "n", reqCount, "Request count")
	pflag.Int64P("concurrency", "c", maxConcurrency, "Number of parrallel requests")
	pflag.IntP("timelimit", "t", benchmarkTimelimit, "Benchmark timeout")
	pflag.IntP("timeout", "s", reqTimeout, "Request timeout")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if len(os.Args) == 1 {
		fmt.Println("Usage of ./spikeys: [options] [http[s]://]hostname[:port]/path")
		pflag.CommandLine.PrintDefaults()
		return
	}
	if flagSet("requests") {
		reqCount = viper.GetInt("requests")
	}
	if flagSet("concurrency") {
		maxConcurrency = viper.GetInt64("concurrency")
	}
	if flagSet("timeout") {
		reqTimeout = viper.GetInt("timeout")
	}
	if flagSet("timelimit") {
		benchmarkTimelimit = viper.GetInt("timelimit")
	}

	positionalArgs := pflag.Args()
	address := positionalArgs[0]
	sem = semaphore.NewWeighted(maxConcurrency)
	results = make([]*result, reqCount)
	u, _ := url.Parse(address)

	fmt.Println("Benchmarking ", u.Hostname(), "...")
	var startTime time.Time
	done := make(chan struct{})
	outCtx, outCancel := context.WithTimeout(context.Background(), time.Duration(benchmarkTimelimit)*time.Second)

	go func() {
		defer func() {
			done <- struct{}{}
		}()
		startTime = time.Now()
		for i := 0; i < reqCount; i++ {
			err := sem.Acquire(outCtx, 1)
			if err != nil {
				sem.Release(1)
				break
			}
			go func(num int) {
				ctx, _ := context.WithTimeout(outCtx, time.Duration(reqTimeout)*time.Second)
				err1 := makeRequest(ctx, address, num)
				if err1 != nil {
					outCancel()
				}
				defer sem.Release(1)
			}(i)
		}

		err := sem.Acquire(outCtx, int64(maxConcurrency))
		if err != nil {
			return
		}
	}()
	select {
	case <-outCtx.Done():
		fmt.Println("The timeout specified has expired")
		outCancel()
		return
	case <-done:
		testDuration = time.Since(startTime)
		PrintResult()
	}
}

func PrintResult() {
	PrintResultItem("Complited requests:", GetComplitedReqCount())
	PrintResultItem("Error requests:", GetErrorReqCount())
	PrintResultItem("Concurrency Level:", maxConcurrency)
	PrintResultItem("Time taken for tests:", fmt.Sprintf("%.3f seconds", testDuration.Seconds()))
	PrintResultItem("Time per request:", fmt.Sprintf("%d [ms] (mean)", GetMeanRequestTime()))
	PrintResultItem("Request per second:", fmt.Sprintf("%.3f [req/sec] (mean)", float64(GetComplitedReqCount())/testDuration.Seconds()))
	PrintResultItem("Total transferred:", fmt.Sprintf("%d bytes", GetTotalTransferredBytes()))
}

func GetTotalTransferredBytes() int64 {
	var bytes int64
	for _, res := range results {
		if res != nil {
			bytes += res.Bytes
		}
	}
	return bytes
}

func GetMeanRequestTime() time.Duration {
	return time.Duration(GetComplitedReqDurationSum() / int64(GetComplitedReqCount()))
}

func GetRequestPerSecond() int64 {
	return int64(GetComplitedReqCount() / GetComplitedReqDurationSeconds())
}

func GetComplitedReqDurationSeconds() int64 {
	var duration float64
	for _, res := range results {
		if res != nil {
			sec := res.Time.Seconds()
			duration += sec
		}
	}
	return int64(duration)
}

func GetComplitedReqDurationSum() int64 {
	var duration time.Duration
	for _, res := range results {
		if res != nil {
			duration += res.Time
		}
	}
	return duration.Milliseconds()
}

func GetComplitedReqCount() int64 {
	var count int64
	for _, res := range results {
		if res != nil {
			count++
		}
	}
	return count
}

func GetErrorReqCount() int64 {
	var count int64
	for _, res := range results {
		if res != nil && res.err != nil {
			count++
		}
	}
	return count
}

func PrintResultItem(item string, value interface{}) {
	fmt.Println(item, MakeSpaces(item), value)
}
func MakeSpaces(item string) string {
	valueStartPosition := 24
	spaceCount := valueStartPosition - len(item)
	result := ""
	for i := 0; i < spaceCount; i++ {
		result += " "
	}
	return result
}
func makeRequest(ctx context.Context, address string, reqNimber int) error {
	var resultChan = make(chan result)

	go func() {
		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, "GET", address, nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			resultChan <- result{time.Since(start), 0, err}
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			resultChan <- result{time.Since(start), 0, err}
			return
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			resultChan <- result{time.Since(start), 0, err}
		}
		duration := time.Since(start)
		resultChan <- result{duration, int64(len(b)), nil}
	}()

	select {
	case <-ctx.Done():
		newResilt := result{time.Duration(reqTimeout), 0, ctx.Err()}
		results[reqNimber] = &newResilt
		return ctx.Err()
	case result := <-resultChan:
		results[reqNimber] = &result
		return nil
	}
}