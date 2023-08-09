package sdkmetrics

import "time"

type RegistryOption func(*realMetricRegistry)

func SetSource(source string) RegistryOption {
	return func(registry *realMetricRegistry) {
		registry.source = source
	}
}

func SetReportTicker(ticker *time.Ticker) RegistryOption {
	return func(registry *realMetricRegistry) {
		registry.reportTicker = ticker
	}
}

func SetTags(tags map[string]string) RegistryOption {
	return func(registry *realMetricRegistry) {
		registry.tags = tags
	}
}

func SetTag(key, value string) RegistryOption {
	return func(registry *realMetricRegistry) {
		if registry.tags == nil {
			registry.tags = make(map[string]string)
		}
		registry.tags[key] = value
	}
}

func SetPrefix(prefix string) RegistryOption {
	return func(registry *realMetricRegistry) {
		registry.prefix = prefix
	}
}
