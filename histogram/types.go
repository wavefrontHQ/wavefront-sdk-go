package histogram

import (
	"sort"
	"time"
)

// Centroid encapsulates a mean value and the count of points associated with that value.
type Centroid struct {
	Value float64
	Count int
}

type Centroids []Centroid

func (a Centroids) Len() int           { return len(a) }
func (a Centroids) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Centroids) Less(i, j int) bool { return a[i].Value < a[j].Value }

func (centroids Centroids) Compact() Centroids {
	res := make(Centroids, 0)
	tmp := make(map[float64]int)
	for _, c := range centroids {
		if _, ok := tmp[c.Value]; ok {
			tmp[c.Value] += c.Count
		} else {
			tmp[c.Value] = c.Count
		}
	}
	for v, c := range tmp {
		res = append(res, Centroid{Value: v, Count: c})
	}
	sort.Sort(res)
	return res
}

// Granularity is the interval (MINUTE, HOUR and/or DAY) by which the histogram data should be aggregated.
type Granularity int8

const (
	MINUTE Granularity = iota
	HOUR
	DAY
)

// Duration of the Granularity
func (hg *Granularity) Duration() time.Duration {
	switch *hg {
	case MINUTE:
		return time.Minute
	case HOUR:
		return time.Hour
	default:
		return time.Hour * 24
	}
}

func (hg *Granularity) String() string {
	switch *hg {
	case MINUTE:
		return "!M"
	case HOUR:
		return "!H"
	default: // DAY
		return "!D"
	}
}
