package utils

import (
	"os"
)

func PtrSliceToDistinct[T comparable](slice []*T) []T {
	_map := make(map[T]struct{})
	newSlice := []T{}

	for _, value := range slice {
		if _, _value := _map[*value]; !_value {
			_map[*value] = struct{}{}
			newSlice = append(newSlice, *value)
		}
	}

	return newSlice
}

func SliceToDistinct[T comparable](slice []T) []T {
	_map := make(map[T]struct{})
	newSlice := []T{}

	for _, value := range slice {
		if _, _value := _map[value]; !_value {
			_map[value] = struct{}{}
			newSlice = append(newSlice, value)
		}
	}

	return newSlice
}

func GetEnvironment() string {
	return os.Getenv("ENVIRONMENT")
}