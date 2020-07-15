package internal

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

var semVerRegex, _ = regexp.Compile("([0-9]\\d*)\\.(\\d+)\\.(\\d+)(?:-([a-zA-Z0-9]+))?")

func GetHostname(defaultVal string) string {
	hostname, err := os.Hostname()
	if err != nil {
		return defaultVal
	}
	return hostname
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GetSemVer(version string) (float64, error) {
	if len(version) > 0 {
		res := semVerRegex.FindStringSubmatch(version)
		major := res[1]
		minor := res[2]
		patch := res[3]
		var sdkVersion strings.Builder
		sdkVersion.WriteString(major)
		sdkVersion.WriteString(".")

		if len(minor) == 1 {
			sdkVersion.WriteString("0" + minor)
		} else {
			sdkVersion.WriteString(minor)
		}

		if len(patch) == 1 {
			sdkVersion.WriteString("0" + patch)
		} else {
			sdkVersion.WriteString(patch)
		}
		return strconv.ParseFloat(sdkVersion.String(), 64)
	}
	return 0, nil
}
