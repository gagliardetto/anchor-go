package internal

func IsKnownBrokenIdl(idlFilename string) bool {
	return isAnyOf(idlFilename)
}

func isAnyOf(value string, values ...string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
