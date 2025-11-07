package helper

import (
	"strconv"
	"strings"
)

func ResolveUnionIntOrStringValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case string:
		return convertByteStringToInt(v)
	default:
		return 0
	}
}

func convertByteStringToInt(s string) int {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	multiplier := 1
	if strings.HasSuffix(s, "kb") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "kb")
	} else if strings.HasSuffix(s, "mb") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	} else if strings.HasSuffix(s, "gb") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	}

	num, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}

	return num * multiplier
}
