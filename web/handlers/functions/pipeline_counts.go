package functions

import (
	"sort"
)

func SortPipelineCounts(counts map[string]int) []map[string]interface{} {
	result := []map[string]interface{}{}
	for key, value := range counts {
		result = append(result, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i]["key"].(string) == "Other" {
			return false
		}
		if result[j]["key"].(string) == "Other" {
			return true
		}
		return result[i]["value"].(int) > result[j]["value"].(int)
	})
	return result
}
