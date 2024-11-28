package benchmarking

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
)

const (
	DefaultRequestCount     = 1
	DefaultMaxConcurrency   = 1
	DefaultBenchmarkTimeout = 30
	DefaultRequestTimeout   = 5
)

type BenchmarkConfig struct {
	RequestCount     int
	MaxConcurrency   int
	RequestTimeout   time.Duration
	BenchmarkTimeout time.Duration
	URL              string
}

func LoadConfig(args []string, flagset *pflag.FlagSet, config *BenchmarkConfig) error {
	var requestCount int
	var maxConcurrency int
	var requestTimeout int
	var benchmarkTimeout int

	flagset.IntVarP(&requestCount, "requests", "n", DefaultRequestCount, "Request count")
	flagset.IntVarP(&maxConcurrency, "concurrency", "c", DefaultMaxConcurrency, "Number of parallel requests")
	flagset.IntVarP(&benchmarkTimeout, "timelimit", "t", DefaultBenchmarkTimeout, "Benchmark timeout (seconds)")
	flagset.IntVarP(&requestTimeout, "timeout", "s", DefaultRequestTimeout, "Request timeout (seconds)")

	if err := flagset.Parse(args[1:]); err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	positionalArgs := flagset.Args()
	if len(positionalArgs) < 1 {
		return fmt.Errorf("missing positional argument: target URL")
	}

	address := positionalArgs[0]
	url, err := url.ParseRequestURI(address)
	if err != nil || url.Scheme == "" || url.Host == "" {
		return fmt.Errorf("invalid URL: %s", address)
	}

	if requestCount <= 0 {
		return fmt.Errorf("invalid request count: must be greater than 0")
	}
	if maxConcurrency <= 0 {
		return fmt.Errorf("invalid concurrency: must be greater than 0")
	}
	if benchmarkTimeout <= 0 {
		return fmt.Errorf("invalid benchmark timeout: must be greater than 0")
	}
	if requestTimeout <= 0 {
		return fmt.Errorf("invalid request timeout: must be greater than 0")
	}

	config.URL = address
	config.BenchmarkTimeout = time.Duration(benchmarkTimeout) * time.Second
	config.RequestTimeout = time.Duration(requestTimeout) * time.Second
	config.MaxConcurrency = maxConcurrency
	config.RequestCount = requestCount

	return nil
}
