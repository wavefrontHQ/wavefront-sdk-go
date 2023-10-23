package lines

import "github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"

type LineHandlerOption func(*RealHandler)

func SetRegistry(registry sdkmetrics.Registry) LineHandlerOption {
	return func(handler *RealHandler) {
		handler.internalRegistry = registry
	}
}

func SetHandlerPrefix(prefix string) LineHandlerOption {
	return func(handler *RealHandler) {
		handler.prefix = prefix
	}
}

func SetLockOnThrottledError(lock bool) LineHandlerOption {
	return func(handler *RealHandler) {
		handler.lockOnErrThrottled = lock
	}
}
