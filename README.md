# Spikeys: Utility for HTTP Load Testing

Spikeys is a simple and efficient tool for testing the performance of HTTP and HTTPS endpoints under load.

---

## **Usage**
```bash
./spikeys [options] [http[s]://]hostname[:port]/path

Options
-c, --concurrency int
Number of parallel requests (default: 1).

-n, --requests int
Total number of requests to send (default: 1).

-t, --timelimit int
Benchmark duration in seconds (default: 30).

-s, --timeout int
Timeout for each request in seconds (default: 30).

