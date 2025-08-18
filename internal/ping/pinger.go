package ping

import (
	"swift-dns/internal/config"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Result struct {
	AvgRtt   time.Duration
	TimedOut bool
}

type JobResult struct {
	Provider config.Provider
	Result   Result
}

func Ping(address string) (Result, error) {
	pinger, err := probing.NewPinger(address)

	if err != nil {
		return Result{}, err
	}

	pinger.SetPrivileged(true)
	pinger.Count = 3
	pinger.Timeout = 3 * time.Second

	err = pinger.Run()

	if err != nil {
		return Result{TimedOut: true}, nil
	}

	stats := pinger.Statistics()

	if stats.PacketsRecv == 0 {
		return Result{TimedOut: true}, nil
	}

	return Result{AvgRtt: stats.AvgRtt, TimedOut: false}, nil
}
