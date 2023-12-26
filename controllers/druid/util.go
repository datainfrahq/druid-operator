package druid

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func firstNonEmptyStr(s1 string, s2 string) string {
	if len(s1) > 0 {
		return s1
	} else {
		return s2
	}
}

// Note that all the arguments passed to this function must have zero value of Nil.
func firstNonNilValue(v1, v2 interface{}) interface{} {
	if !reflect.ValueOf(v1).IsNil() {
		return v1
	} else {
		return v2
	}
}

// lookup DENY_LIST, default is nil
func getDenyListEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// pass slice of strings for namespaces
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getDenyListEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	// split on ","
	val := strings.Split(valStr, sep)
	return val
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// returns pointer to bool
func boolFalse() *bool {
	bool := false
	return &bool
}

// to be used in max concurrent reconciles only
// defaulting to return 1
func Str2Int(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return i
}

func IsEqualJson(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
