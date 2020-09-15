package utils

func ContainsString(items []string, item string) bool {
	for _, it := range items {
		if it == item {
			return true
		}
	}
	return false
}
