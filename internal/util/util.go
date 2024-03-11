package util

import "encoding/json"

func JsonPrint(i interface{}) string {
	s, _ := json.Marshal(i)
	return string(s)
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
