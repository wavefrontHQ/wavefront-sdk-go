package internal

func IsUUIDFormat(str string) bool {
	l := len(str)
	if l != 36 {
		return false
	}
	for i := 0; i < l; i++ {
		c := str[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')) {
			return false
		}
	}
	return true
}
