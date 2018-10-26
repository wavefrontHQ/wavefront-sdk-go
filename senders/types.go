package senders

// A centroid encapsulates a mean value and the count of points associated with that value.
type Centroid struct {
	Value float64
	Count int
}

// The interval (MINUTE, HOUR and/or DAY) by which the histogram data should be aggregated.
type HistogramGranularity int8

const (
	MINUTE HistogramGranularity = iota
	HOUR
	DAY
)

func (hg *HistogramGranularity) String() string {
	switch *hg {
	case MINUTE:
		return "!M"
	case HOUR:
		return "!H"
	case DAY:
		return "!D"
	}
	return ""
}

type SpanTag struct {
	Key   string
	Value string
}

type SpanLog struct {
	Timestamp int64
	Fields    map[string]string
}
