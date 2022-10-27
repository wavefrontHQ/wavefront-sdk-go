package event

func AdjustStartEndTime(startMillis, endMillis int64) (int64, int64) {
	// secs to millis
	if startMillis < 999999999999 {
		startMillis = startMillis * 1000
	}

	if endMillis <= 999999999999 {
		endMillis = endMillis * 1000
	}

	if endMillis == 0 {
		endMillis = startMillis + 1
	}
	return startMillis, endMillis
}
