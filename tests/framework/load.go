package framework

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

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
		return fmt.Errorf("Error getting proxy URL: %w", err)
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

	err = reporter.Report(outFile)
	if err != nil {
		return fmt.Errorf("Error reporting results: %w", err)
	}
	return nil
}
