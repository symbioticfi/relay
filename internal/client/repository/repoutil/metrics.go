package repoutil

import "time"

type Metrics interface {
	ObserveRepoQueryDuration(queryName string, status string, d time.Duration)
	ObserveRepoQueryTotalDuration(queryName string, status string, d time.Duration)
	SetDBSizeBytes(sizeBytes float64)
}

type DoNothingMetrics struct{}

func (DoNothingMetrics) ObserveRepoQueryDuration(string, string, time.Duration)      {}
func (DoNothingMetrics) ObserveRepoQueryTotalDuration(string, string, time.Duration) {}
func (DoNothingMetrics) SetDBSizeBytes(float64)                                      {}
