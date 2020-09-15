package util

import (
	"github.com/ghodss/yaml"
)

// ToObjectMap converts the given object into a map of strings/maps using YAML marshalling
func ToObjectMap(object interface{}) (map[string]interface{}, error) {
	answer := map[string]interface{}{}
	data, err := yaml.Marshal(object)
	if err != nil {
		return answer, err
	}
	err = yaml.Unmarshal(data, &answer)
	return answer, err
}
