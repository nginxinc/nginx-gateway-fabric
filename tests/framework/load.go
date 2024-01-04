package framework

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Target struct {
	Header http.Header
	Method string
	URL    string
	Body   []byte
}

func convertTargetToVegetaTarget(targets []Target) []vegeta.Target {
	vegTargets := make([]vegeta.Target, 0, len(targets))
	for _, t := range targets {
		vt := vegeta.Target{
			Method: t.Method,
			URL:    t.URL,
			Body:   t.Body,
			Header: t.Header,
		}
		vegTargets = append(vegTargets, vt)
	}
	return vegTargets
}

// Metrics is a wrapper around the vegeta Metrics.
type Metrics struct {
	vegeta.Metrics
}

// RunLoadTest uses Vegeta to send traffic to the provided Targets at the given rate for the given duration and writes
// the results to the provided file
func RunLoadTest(
	targets []Target,
	rate int,
	duration time.Duration,
	desc,
	proxy,
	serverName string,
) (vegeta.Results, Metrics) {
	vegTargets := convertTargetToVegetaTarget(targets)
	targeter := vegeta.NewStaticTargeter(vegTargets...)

	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: vegeta.DefaultLocalAddr.IP, Zone: vegeta.DefaultLocalAddr.Zone},
		KeepAlive: 30 * time.Second,
	}

	httpClient := http.Client{
		Timeout: vegeta.DefaultTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, proxy)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // self-signed cert for testing
				ServerName:         serverName,
			},
			MaxIdleConnsPerHost: vegeta.DefaultConnections,
			MaxConnsPerHost:     vegeta.DefaultMaxConnections,
		},
	}

	attacker := vegeta.NewAttacker(vegeta.Client(&httpClient))

	r := vegeta.Rate{Freq: rate, Per: time.Second}
	var results vegeta.Results
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, r, duration, desc) {
		results = append(results, *res)
		metrics.Add(res)
	}
	metrics.Close()

	return results, Metrics{metrics}
}
