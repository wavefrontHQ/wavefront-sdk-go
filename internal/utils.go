package internal

import "os"

func GetHostnmae(defaultVal string) string {
	hostname, err := os.Hostname()
	if err != nil {
		return defaultVal
	}
	return hostname
}
