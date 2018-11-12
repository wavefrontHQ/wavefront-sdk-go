package wavefront

import (
	"log"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/senders"

	tdigest "github.com/caio/go-tdigest"
)

type Histogram struct {
	mutex              sync.Mutex
	priorTimedBinsList []*timedBin
	currentTimedBin    *timedBin

	metricName string
	source     string
	tags       map[string]string

	Granularity senders.HistogramGranularity
	Compression uint32
	MaxBins     int
}

type timedBin struct {
	tdigest   *tdigest.TDigest
	timestamp time.Time
}

type distribution struct {
	centroids []senders.Centroid
	timestamp time.Time
}

func NewHistogram(metricName string, source string, tags map[string]string) *Histogram {
	h := &Histogram{}
	h.MaxBins = 10
	h.Granularity = senders.MINUTE
	h.Compression = 5

	h.metricName = metricName
	h.source = source
	h.tags = tags

	return h
}

func (h *Histogram) Add(v float64) {
	h.rotateCurrentTDigestIfNeedIt()
	h.currentTimedBin.tdigest.Add(v)
}

func (h *Histogram) Report(sender senders.Sender) {
	distributions := h.getDistributions()
	log.Printf("metric: '%s' count:'%d'", h.metricName, len(distributions))
	hgs := map[senders.HistogramGranularity]bool{h.Granularity: true}
	for _, distribution := range distributions {
		sender.SendDistribution(h.metricName, distribution.centroids, hgs, distribution.timestamp.Unix(), h.source, h.tags)
	}
}

func (h *Histogram) getDistributions() []distribution {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	distributions := make([]distribution, len(h.priorTimedBinsList))
	for idx, bin := range h.priorTimedBinsList {
		var centroids []senders.Centroid
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			centroids = append(centroids, senders.Centroid{Value: mean, Count: int(count)})
			return true
		})
		distributions[idx] = distribution{timestamp: bin.timestamp, centroids: centroids}
	}
	h.priorTimedBinsList = h.priorTimedBinsList[:0]
	return distributions
}

func (h *Histogram) rotateCurrentTDigestIfNeedIt() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.currentTimedBin == nil {
		h.currentTimedBin = h.newTimedBin()
	} else if h.currentTimedBin.timestamp != h.now() {
		h.priorTimedBinsList = append(h.priorTimedBinsList, h.currentTimedBin)
		if len(h.priorTimedBinsList) > h.MaxBins {
			h.priorTimedBinsList = h.priorTimedBinsList[1:]
		}
		h.currentTimedBin = h.newTimedBin()
	}
}

func (h *Histogram) now() time.Time {
	switch h.Granularity {
	case senders.HOUR:
		return time.Now().Truncate(time.Hour)
	case senders.DAY:
		return time.Now().Truncate(time.Hour * 24)
	default:
		return time.Now().Truncate(time.Minute)
	}
}

func (h *Histogram) newTimedBin() *timedBin {
	td, _ := tdigest.New(tdigest.Compression(h.Compression))
	return &timedBin{timestamp: h.now(), tdigest: td}
}
