package mapper

import "strconv"

func Atoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return 0
	}

	return v
}

func AtoiWithDefault(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return def
	}
	return i
}
