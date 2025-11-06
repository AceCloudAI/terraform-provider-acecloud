package helpers

import (
	"fmt"
)


func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	
	errorStr := err.Error()
	return containsAny(errorStr, []string{
		"not found",
		"404",
		"does not exist",
		"resource not found",
		"VM not found",
	})
}


func ConvertToInt(val interface{}) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}


func InterfaceSliceToStringSlice(ifaceSlice []interface{}) []string {
	strSlice := make([]string, len(ifaceSlice))
	for i, v := range ifaceSlice {
		strSlice[i] = v.(string)
	}
	return strSlice
}


func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}