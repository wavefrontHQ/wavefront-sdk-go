package event

// Annotation field configuration
type Annotation func(map[string]string)

// Severity set the envet 'severity' annotation
func Severity(severity string) Annotation {
	return func(annotations map[string]string) {
		annotations["severity"] = severity
	}
}

// Type set the envet 'type' annotation
func Type(t string) Annotation {
	return func(annotations map[string]string) {
		annotations["type"] = t
	}
}

// Details set the envet 'details' annotation
func Details(details string) Annotation {
	return func(annotations map[string]string) {
		annotations["details"] = details
	}
}
