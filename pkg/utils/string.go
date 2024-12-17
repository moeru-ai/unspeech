package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var (
	errFailedToConvertStringToType = func(t any, err error) error { return fmt.Errorf("failed to convert string to type %T: %w", t, err) }
)

func FromString[T any](str string) (T, error) { //nolint:gocyclo
	var empty T
	if str == "" {
		switch any(empty).(type) {
		case []byte:
			val, _ := any(make([]byte, 0)).(T)
			return val, nil
		case []rune:
			val, _ := any(make([]rune, 0)).(T)
			return val, nil
		case *strings.Builder:
			val, _ := any(&strings.Builder{}).(T)
			return val, nil
		}

		return empty, nil
	}
	if str == "null" {
		return empty, nil
	}
	if str == "<nil>" {
		return empty, nil
	}

	switch any(empty).(type) {
	case string:
		val, _ := any(str).(T)
		return val, nil
	case int:
		val, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(int(val)).(T)

		return typeVal, nil
	case int8:
		val, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(int8(val)).(T)

		return typeVal, nil
	case int16:
		val, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(int16(val)).(T)

		return typeVal, nil
	case int32:
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(int32(val)).(T)

		return typeVal, nil
	case int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(val).(T)

		return typeVal, nil
	case uint:
		val, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(uint(val)).(T)

		return typeVal, nil
	case uint8:
		val, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(uint8(val)).(T)

		return typeVal, nil
	case uint16:
		val, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(uint16(val)).(T)

		return typeVal, nil
	case uint32:
		val, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(uint32(val)).(T)

		return typeVal, nil
	case uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(val).(T)

		return typeVal, nil
	case float32:
		val, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(float32(val)).(T)

		return typeVal, nil
	case float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(val).(T)

		return typeVal, nil
	case complex64:
		val, err := strconv.ParseComplex(str, 64)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(complex64(val)).(T)

		return typeVal, nil
	case complex128:
		val, err := strconv.ParseComplex(str, 128)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(val).(T)

		return typeVal, nil
	case bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		typeVal, _ := any(val).(T)

		return typeVal, nil
	case []byte:
		val, _ := any([]byte(str)).(T)
		return val, nil
	case []rune:
		val, _ := any([]rune(str)).(T)
		return val, nil
	case *strings.Builder:
		var sb strings.Builder

		sb.WriteString(str)
		val, _ := any(&sb).(T)

		return val, nil
	default:
		var initial T

		err := json.Unmarshal([]byte(str), &initial)
		if err != nil {
			return empty, errFailedToConvertStringToType(empty, err)
		}

		return initial, nil
	}
}

func FromStringOrEmpty[T any](str string) T {
	var empty T

	val, err := FromString[T](str)
	if err != nil {
		return empty
	}

	return val
}

func IsNumber(str string) bool {
	_, err := strconv.ParseFloat(str, 64)

	return err == nil
}
