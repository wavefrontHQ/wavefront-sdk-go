package histogram

import (
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/senders"

	tdigest "github.com/caio/go-tdigest"
)

type Histogram struct {
	mutex              sync.Mutex
	priorTimedBinsList []*timedBin
	currentTimedBin    *timedBin

	Granularity time.Duration
	Compression uint32
	MaxBins     int
}

type timedBin struct {
	tdigest   *tdigest.TDigest
	timestamp time.Time
}

type Distribution struct {
	Centroids []senders.Centroid
	Timestamp time.Time
}

func NewHistogram() *Histogram {
	h := &Histogram{MaxBins: 10, Granularity: time.Minute, Compression: 5}
	return h
}

func (h *Histogram) Add(v float64) {
	h.rotateCurrentTDigestIfNeedIt()
	h.currentTimedBin.tdigest.Add(v)
}

func (h *Histogram) Distributions() []Distribution {
	h.rotateCurrentTDigestIfNeedIt()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	distributions := make([]Distribution, len(h.priorTimedBinsList))
	for idx, bin := range h.priorTimedBinsList {
		var centroids []senders.Centroid
		bin.tdigest.ForEachCentroid(func(mean float64, count uint32) bool {
			centroids = append(centroids, senders.Centroid{Value: mean, Count: int(count)})
			return true
		})
		distributions[idx] = Distribution{Timestamp: bin.timestamp, Centroids: centroids}
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
	return time.Now().Truncate(h.Granularity)
}

func (h *Histogram) newTimedBin() *timedBin {
	td, _ := tdigest.New(tdigest.Compression(h.Compression))
	return &timedBin{timestamp: h.now(), tdigest: td}
}
