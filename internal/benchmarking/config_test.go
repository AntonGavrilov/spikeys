package benchmarking

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigURLErr(t *testing.T) {

	config := &BenchmarkConfig{}
	args1 := []string{""}
	flagSet := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)

	err := LoadConfig(args1, flagSet, config)
	require.ErrorContains(t, err, "missing positional argument: target URL")
	assert.Empty(t, config)
}

func TestLoadConfigInvalidURLErr(t *testing.T) {

	config := &BenchmarkConfig{}
	args1 := []string{"file", "invalidurl.com"}
	flagSet := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)
	err := LoadConfig(args1, flagSet, config)
	require.ErrorContains(t, err, "invalid URL:")
	assert.Empty(t, config)

}

func TestLoadConfigDefaultParametersSetting(t *testing.T) {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	config := &BenchmarkConfig{}
	args1 := []string{"file", "http://validUrl.com"}
	flagSet := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)

	err := LoadConfig(args1, flagSet, config)

	require.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, time.Duration(DefaultBenchmarkTimeout)*time.Second, config.BenchmarkTimeout)
	assert.Equal(t, time.Duration(DefaultRequestTimeout)*time.Second, config.RequestTimeout)
	assert.Equal(t, DefaultMaxConcurrency, config.MaxConcurrency)
	assert.Equal(t, DefaultRequestCount, config.RequestCount)
}

func TestLoadBecnhmarkConfigParseArgs(t *testing.T) {
	tests := []struct {
		name           string          
		args           []string        
		expectedError  string          
		expectedConfig BenchmarkConfig 
	}{
		{
			name: "Valid input with all flags",
			args: []string{
				"bench",
				"--requests=50",
				"--concurrency=5",
				"--timelimit=120",
				"--timeout=10",
				"https://example.com",
			},
			expectedError: "",
			expectedConfig: BenchmarkConfig{
				URL:              "https://example.com",
				BenchmarkTimeout: 120 * time.Second,
				RequestTimeout:   10 * time.Second,
				MaxConcurrency:   5,
				RequestCount:     50,
			},
		},
		{
			name: "Invalid request count",
			args: []string{
				"bench",
				"--requests=0",
				"https://example.com",
			},
			expectedError: "invalid request count: must be greater than 0",
		},
		{
			name: "Invalid concurrency",
			args: []string{
				"bench",
				"--concurrency=0",
				"https://example.com",
			},
			expectedError: "invalid concurrency: must be greater than 0",
		},
		{
			name: "Invalid benchmark timeout",
			args: []string{
				"bench",
				"--timelimit=0",
				"https://example.com",
			},
			expectedError: "invalid benchmark timeout: must be greater than 0",
		},
		{
			name: "Invalid Request timeout",
			args: []string{
				"bench",
				"--timeout=0",
				"https://example.com",
			},
			expectedError: "invalid request timeout: must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var config BenchmarkConfig
			flagSet := pflag.NewFlagSet("benchmark", pflag.ContinueOnError)

			err := LoadConfig(tt.args, flagSet, &config)

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, config)
			}
		})
	}
}
