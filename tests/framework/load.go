package framework

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// RunLoadTest uses Vegeta to send traffic to the provided Targets at the given rate for the given duration and writes
// the results to the provided file
func RunLoadTest(
	targets []vegeta.Target,
	rate int,
	duration time.Duration,
	desc string,
	outFile *os.File,
	proxy string,
) error {
	targeter := vegeta.NewStaticTargeter(targets...)
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return fmt.Errorf("error getting proxy URL: %w", err)
	}

	attacker := vegeta.NewAttacker(
		vegeta.Proxy(http.ProxyURL(proxyURL)),
	)

	r := vegeta.Rate{Freq: rate, Per: time.Second}
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, r, duration, desc) {
		metrics.Add(res)
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)

	if err = reporter.Report(outFile); err != nil {
		return fmt.Errorf("error reporting results: %w", err)
	}
	return nil
}
