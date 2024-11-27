package benchmarking

import "errors"

var ErrRequestTimeoutExceeded = errors.New("request Timeout exceeded")

var ErrBenchmarkTimeoutExcceded = errors.New("benchmark Timeout exceeded")

var ErrInvalidConfig = errors.New("invalid config")
