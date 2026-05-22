package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RemediationAttemptsTotal tracks how many times sentinel has tried to fix things.
	RemediationAttemptsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sentinel_remediation_attempts_total",
		Help: "The total number of automated remediation attempts.",
	})

	// ActiveLockouts tracks the current number of services locked by the circuit breaker.
	ActiveLockouts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sentinel_active_lockouts",
		Help: "The current number of active circuit breaker lockouts.",
	})
)
