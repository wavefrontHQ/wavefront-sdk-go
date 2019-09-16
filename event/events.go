package event

// Option configuration
type Option func(map[string]interface{})

// Severity set the envet 'severity' annotation
func Severity(severity string) Option {
	return func(event map[string]interface{}) {
		annotations := event["annotations"].(map[string]string)
		annotations["severity"] = severity
	}
}

// Type set the envet 'type' annotation
func Type(t string) Option {
	return func(event map[string]interface{}) {
		annotations := event["annotations"].(map[string]string)
		annotations["type"] = t
	}
}

// Details set the envet 'details' annotation
func Details(details string) Option {
	return func(event map[string]interface{}) {
		annotations := event["annotations"].(map[string]string)
		annotations["details"] = details
	}
}
