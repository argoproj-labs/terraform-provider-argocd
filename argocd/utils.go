package argocd

import "strconv"

func convertStringToInt64(s string) (i int64, err error) {
	i, err = strconv.ParseInt(s, 10, 64)
	return
}

func convertInt64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func isKeyInMap(key string, d map[string]interface{}) bool {
	if d == nil {
		return false
	}
	for k := range d {
		if k == key {
			return true
		}
	}
	return false
}

func expandStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}
