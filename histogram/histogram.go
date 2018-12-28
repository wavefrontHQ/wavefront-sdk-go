package histogram

import (
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/senders"

	tdigest "github.com/caio/go-tdigest"
)

type Histogram interface {
	Update(v float64)
	Distributions() []Distribution
}

type HistogramOption func(*histogramImpl)

func Granularity(g time.Duration) HistogramOption {
	return func(args *histogramImpl) {
		args.granularity = g
	}
}

func Compression(c uint32) HistogramOption {
	return func(args *histogramImpl) {
		args.compression = c
	}
}

func MaxBins(c int) HistogramOption {
	return func(args *histogramImpl) {
		args.maxBins = c
	}
}

func defaultHistogramImpl() *histogramImpl {
	return &histogramImpl{maxBins: 10, granularity: time.Minute, compression: 5}
}

func NewHistogram() Histogram {
	return defaultHistogramImpl()
}

func NewHistogramWithOptions(setters ...HistogramOption) Histogram {
	h := defaultHistogramImpl()
	for _, setter := range setters {
		setter(h)
	}
	return h
}

type histogramImpl struct {
	mutex              sync.Mutex
	priorTimedBinsList []*timedBin
	currentTimedBin    *timedBin

	granularity time.Duration
	compression uint32
	maxBins     int
}

type timedBin struct {
	tdigest   *tdigest.TDigest
	timestamp time.Time
}

type Distribution struct {
	Centroids []senders.Centroid
	Timestamp time.Time
}

func (h *histogramImpl) Update(v float64) {
	h.rotateCurrentTDigestIfNeedIt()
	h.currentTimedBin.tdigest.Add(v)
}

func (h *histogramImpl) Distributions() []Distribution {
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

func (h *histogramImpl) rotateCurrentTDigestIfNeedIt() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.currentTimedBin == nil {
		h.currentTimedBin = h.newTimedBin()
	} else if h.currentTimedBin.timestamp != h.now() {
		h.priorTimedBinsList = append(h.priorTimedBinsList, h.currentTimedBin)
		if len(h.priorTimedBinsList) > h.maxBins {
			h.priorTimedBinsList = h.priorTimedBinsList[1:]
		}
		h.currentTimedBin = h.newTimedBin()
	}
}

func (h *histogramImpl) now() time.Time {
	return time.Now().Truncate(h.granularity)
}

func (h *histogramImpl) newTimedBin() *timedBin {
	td, _ := tdigest.New(tdigest.Compression(h.compression))
	return &timedBin{timestamp: h.now(), tdigest: td}
}
