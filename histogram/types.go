package histogram

import "time"

// A centroid encapsulates a mean value and the count of points associated with that value.
type Centroid struct {
	Value float64
	Count int
}

// The interval (MINUTE, HOUR and/or DAY) by which the histogram data should be aggregated.
type HistogramGranularity int8

const (
	SECOND HistogramGranularity = iota
	MINUTE
	HOUR
	DAY
)

func (hg *HistogramGranularity) Duration() time.Duration {
	switch *hg {
	case SECOND: // just for testing
		return time.Second
	case MINUTE:
		return time.Minute
	case HOUR:
		return time.Hour
	default:
		return time.Hour * 24
	}
}

func (hg *HistogramGranularity) String() string {
	switch *hg {
	case MINUTE:
		return "!M"
	case HOUR:
		return "!H"
	default: // DAY
		return "!D"
	}
}
