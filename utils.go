package main

import (
	"fmt"
	"os"
	"strconv"
)

func AskInputOrEnv[T any](key string) T {
	var val T
	var envValue string

	if v, ok := os.LookupEnv(key); ok {
		envValue = v
	}

	switch any(val).(type) {
	case string:
		if envValue != "" {
			return any(envValue).(T)
		}
	case int:
		if envValue != "" {
			i, err := strconv.Atoi(envValue)
			if err == nil {
				return any(i).(T)
			}
		}
	case int32:
		if envValue != "" {
			i, err := strconv.Atoi(envValue)
			if err == nil {
				return any(int32(i)).(T)
			}
		}
	}

	fmt.Printf("Enter %s: ", key)
	var input string
	fmt.Scanln(&input)

	switch any(val).(type) {
	case string:
		return any(input).(T)
	case int:
		i, err := strconv.Atoi(input)
		if err == nil {
			return any(i).(T)
		}
	}

	return val
}
