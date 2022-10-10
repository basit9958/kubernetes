package utils

import "os"

func Exists(fileName string) bool {
	info, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func Unique(input []string) []string {
	m := make(map[string]bool)
	result := make([]string, 0, len(input))
	for _, s := range input {
		if _, ok := m[s]; !ok {
			m[s] = true
			result = append(result, s)
		}
	}
	return result
}
