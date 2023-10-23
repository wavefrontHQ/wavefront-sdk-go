package lines

// FormatSpecificBatchSender
type Handler interface {
	HandleLine(line string) error
	Start()
	Stop()
	Flush() error
	FlushAll() error
	GetFailureCount() int64
}
