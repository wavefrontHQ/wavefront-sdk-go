package wavefront

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

// Option Wavefront client configuration options
type Option func(*senders.DirectConfiguration)

// NewClient creates Wavefront sender
func NewClient(wfURL string, setters ...Option) (senders.Sender, error) {
	cfg := &senders.DirectConfiguration{}

	u, err := url.Parse(wfURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(strings.ToLower(u.Scheme), "http") {
		return nil, fmt.Errorf("invalid schema '%s', only 'http' is supported", u.Scheme)
	}

	if len(u.User.String()) > 0 {
		cfg.Token = u.User.String()
		u.User = nil
	}

	cfg.Server = u.String()

	for _, set := range setters {
		set(cfg)
	}
	return senders.NewDirectSender(cfg)
}

// BatchSize set max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
func BatchSize(n int) Option {
	return func(cfg *senders.DirectConfiguration) {
		cfg.BatchSize = n
	}
}

// MaxBufferSize set the size of internal buffers beyond which received data is dropped.
func MaxBufferSize(n int) Option {
	return func(cfg *senders.DirectConfiguration) {
		cfg.MaxBufferSize = n
	}
}

// FlushIntervalSeconds set the interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
func FlushIntervalSeconds(n int) Option {
	return func(cfg *senders.DirectConfiguration) {
		cfg.FlushIntervalSeconds = n
	}
}
