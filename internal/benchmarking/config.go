package benchmarking

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/pflag"
)

const (
	DefaultRequestCount     = 100
	DefaultMaxConcurrency   = 10
	DefaultBenchmarkTimeout = 30
	DefaultRequestTimeout   = 5
)

type BecnhmarkConfig struct {
	RequestCount     int
	MaxConcurrency   int
	RequestTimeout   time.Duration
	BenchmarkTimeout time.Duration
	URL              string
}

func NewConfig() BecnhmarkConfig {
	return BecnhmarkConfig{}
}

func LoadConfig(config *BecnhmarkConfig) error {
	var requestCount int
	var maxConcurrency int
	var requestTimeout int
	var benchmarkTimeout int
	pflag.IntVarP(&requestCount, "requests", "n", DefaultRequestCount, "Request count")
	pflag.IntVarP(&maxConcurrency, "concurrency", "c", DefaultMaxConcurrency, "Number of parrallel requests")
	pflag.IntVarP(&benchmarkTimeout, "timelimit", "t", DefaultBenchmarkTimeout, "Benchmark timeout")
	pflag.IntVarP(&requestTimeout, "timeout", "s", DefaultRequestTimeout, "Request timeout")
	pflag.CommandLine.Init("bench", pflag.ContinueOnError)

	if err := pflag.CommandLine.Parse(os.Args); err != nil {
		return err
	}
	positionalArgs := pflag.Args()

	address := positionalArgs[1]

	url, err := url.ParseRequestURI(address)

	if err != nil || url.Scheme == "" || url.Host == "" {
		return fmt.Errorf("invalid url value : %s", address)
	}
	if requestCount == 0 {
		return fmt.Errorf("invalid request count value)")
	}
	if maxConcurrency == 0 {
		return fmt.Errorf("invalid request concurrency value")
	}
	if benchmarkTimeout == 0 {
		return fmt.Errorf("invalid request timelimit value")
	}
	if requestCount == 0 {
		return fmt.Errorf("invalid request timeout value")
	}
	config.URL = address
	config.BenchmarkTimeout = time.Duration(benchmarkTimeout * int(time.Second))
	config.RequestTimeout = time.Duration(benchmarkTimeout * int(time.Second))
	config.MaxConcurrency = maxConcurrency
	config.RequestCount = requestCount
	return err
}
