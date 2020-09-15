package utils

func ToFloat(value interface{}) (float64, bool) {
	if value, ok := value.(float64); ok {
		return value, true
	}
	if value, ok := value.(int); ok {
		return float64(value), true
	}
	return 0, false
}
