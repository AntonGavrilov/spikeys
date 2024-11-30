package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/AntonGavrilov/spikeys/internal/benchmarking"
	"github.com/AntonGavrilov/spikeys/internal/reporting"
	"github.com/spf13/pflag"
)

func main() {
	config := benchmarking.BenchmarkConfig{}
	flagSet := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)
	err := benchmarking.LoadConfig(os.Args, flagSet, &config)

	if err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			fmt.Println(err)
			flagSet.PrintDefaults()
		}
		os.Exit(1)
	}
	benchmarker := benchmarking.NewBenchmarker(
		benchmarking.DefaultRequestPerformer{
			HTTPClient: &http.Client{}})
	result, err := benchmarker.Run(config)

	if err != nil {
		if errors.Is(err, benchmarking.ErrBenchmarkTimeoutExcceded) ||
			errors.Is(err, benchmarking.ErrRequestTimeoutExceeded) {
			fmt.Println(err)
		} else {
			fmt.Println("Unexpecting error during benchmarking")
		}
		os.Exit(1)
	}

	summary := reporting.GenerateSummary(result.Results, result.Duration, int64(config.MaxConcurrency))
	reporting.PrintSummary(summary)
}
