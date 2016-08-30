package api

import (
	"log"
	"os"
	"strconv"
	"strings"
)

func EnvToConfig() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) < 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result
}

func StringToInt(in string) int {
	num, err := strconv.ParseInt(strings.TrimSpace(in), 10, 64)
	if err != nil {
		log.Printf("failed to convert %q to integer", in)
		return 0
	}
	return int(num)
}
