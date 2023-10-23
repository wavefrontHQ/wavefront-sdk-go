package lines


func NewHandler(reporter Reporter) Handler {
	opts := []LineHandlerOption{SetHandlerPrefix(prefix), SetRegistry(registry)}
	batchSize := cfg.BatchSize
	if format == EventFormat {
		batchSize = 1
		opts = append(opts, SetLockOnThrottledError(true))
	}

	return NewLineHandler(reporter, format, cfg.FlushInterval, batchSize, cfg.MaxBufferSize, opts...)
}
