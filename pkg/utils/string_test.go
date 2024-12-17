package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromString(t *testing.T) {
	t.Run("Unsupported", func(t *testing.T) {
		funcVal, err := FromString[func()]("")
		require.NoError(t, err)
		assert.Nil(t, funcVal)

		mapVal, err := FromString[map[string]any]("")
		require.NoError(t, err)
		assert.Nil(t, mapVal)

		mapVal, err = FromString[map[string]any]("")
		require.NoError(t, err)
		assert.Empty(t, mapVal)

		sliceVal, err := FromString[[]string]("")
		require.NoError(t, err)
		assert.Nil(t, sliceVal)

		sliceVal, err = FromString[[]string]("")
		require.NoError(t, err)
		assert.Empty(t, sliceVal)

		structVal, err := FromString[struct{}]("")
		require.NoError(t, err)
		assert.Empty(t, structVal)

		funcVal, err = FromString[func()]("{}")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type func(): json: cannot unmarshal object into Go value of type func()")
		assert.Nil(t, funcVal)
	})

	t.Run("Empty", func(t *testing.T) {
		stringVal, err := FromString[string]("")
		require.NoError(t, err)
		assert.Equal(t, "", stringVal)

		intVal, err := FromString[int]("")
		require.NoError(t, err)
		assert.Zero(t, intVal)

		int8Val, err := FromString[int8]("")
		require.NoError(t, err)
		assert.Zero(t, int8Val)

		int16Val, err := FromString[int16]("")
		require.NoError(t, err)
		assert.Zero(t, int16Val)

		int32Val, err := FromString[int32]("")
		require.NoError(t, err)
		assert.Zero(t, int32Val)

		int64Val, err := FromString[int64]("")
		require.NoError(t, err)
		assert.Zero(t, int64Val)

		uintVal, err := FromString[uint]("")
		require.NoError(t, err)
		assert.Zero(t, uintVal)

		uint8Val, err := FromString[uint8]("")
		require.NoError(t, err)
		assert.Zero(t, uint8Val)

		uint16Val, err := FromString[uint16]("")
		require.NoError(t, err)
		assert.Zero(t, uint16Val)

		uint32Val, err := FromString[uint32]("")
		require.NoError(t, err)
		assert.Zero(t, uint32Val)

		uint64Val, err := FromString[uint64]("")
		require.NoError(t, err)
		assert.Zero(t, uint64Val)

		float32Val, err := FromString[float32]("")
		require.NoError(t, err)
		assert.Zero(t, float32Val)

		float64Val, err := FromString[float64]("")
		require.NoError(t, err)
		assert.Zero(t, float64Val)

		complex64Val, err := FromString[complex64]("")
		require.NoError(t, err)
		assert.Zero(t, complex64Val)

		complex128Val, err := FromString[complex128]("")
		require.NoError(t, err)
		assert.Zero(t, complex128Val)

		boolVal, err := FromString[bool]("")
		require.NoError(t, err)
		assert.False(t, boolVal)

		bytesVal, err := FromString[[]byte]("")
		require.NoError(t, err)
		assert.Empty(t, bytesVal)

		runesVal, err := FromString[[]rune]("")
		require.NoError(t, err)
		assert.Empty(t, runesVal)

		mapVal, err := FromString[map[string]any]("{}")
		require.NoError(t, err)
		assert.Empty(t, mapVal)

		sliceVal, err := FromString[[]string]("[]")
		require.NoError(t, err)
		assert.Empty(t, sliceVal)

		structVal, err := FromString[struct{}]("{}")
		require.NoError(t, err)
		assert.Empty(t, structVal)

		builderVal, err := FromString[*strings.Builder]("")
		require.NoError(t, err)
		assert.NotNil(t, builderVal)

		anyVal, err := FromString[any]("")
		require.NoError(t, err)
		assert.Equal(t, "<nil>", fmt.Sprintf("%T", anyVal))
		assert.Nil(t, anyVal)
	})

	t.Run("Invalid", func(t *testing.T) {
		intVal, err := FromString[int]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type int: strconv.ParseInt: parsing \"invalid\": invalid syntax")
		assert.Zero(t, intVal)

		int8Val, err := FromString[int8]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type int8: strconv.ParseInt: parsing \"invalid\": invalid syntax")
		assert.Zero(t, int8Val)

		int16Val, err := FromString[int16]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type int16: strconv.ParseInt: parsing \"invalid\": invalid syntax")
		assert.Zero(t, int16Val)

		int32Val, err := FromString[int32]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type int32: strconv.ParseInt: parsing \"invalid\": invalid syntax")
		assert.Zero(t, int32Val)

		int64Val, err := FromString[int64]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type int64: strconv.ParseInt: parsing \"invalid\": invalid syntax")
		assert.Zero(t, int64Val)

		uintVal, err := FromString[uint]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type uint: strconv.ParseUint: parsing \"invalid\": invalid syntax")
		assert.Zero(t, uintVal)

		uint8Val, err := FromString[uint8]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type uint8: strconv.ParseUint: parsing \"invalid\": invalid syntax")
		assert.Zero(t, uint8Val)

		uint16Val, err := FromString[uint16]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type uint16: strconv.ParseUint: parsing \"invalid\": invalid syntax")
		assert.Zero(t, uint16Val)

		uint32Val, err := FromString[uint32]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type uint32: strconv.ParseUint: parsing \"invalid\": invalid syntax")
		assert.Zero(t, uint32Val)

		uint64Val, err := FromString[uint64]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type uint64: strconv.ParseUint: parsing \"invalid\": invalid syntax")
		assert.Zero(t, uint64Val)

		float32Val, err := FromString[float32]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type float32: strconv.ParseFloat: parsing \"invalid\": invalid syntax")
		assert.Zero(t, float32Val)

		float64Val, err := FromString[float64]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type float64: strconv.ParseFloat: parsing \"invalid\": invalid syntax")
		assert.Zero(t, float64Val)

		complex64Val, err := FromString[complex64]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type complex64: strconv.ParseComplex: parsing \"invalid\": invalid syntax")
		assert.Zero(t, complex64Val)

		complex128Val, err := FromString[complex128]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type complex128: strconv.ParseComplex: parsing \"invalid\": invalid syntax")
		assert.Zero(t, complex128Val)

		boolVal, err := FromString[bool]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type bool: strconv.ParseBool: parsing \"invalid\": invalid syntax")
		assert.False(t, boolVal)

		mapVal, err := FromString[map[string]any]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type map[string]interface {}: invalid character 'i' looking for beginning of value")
		assert.Nil(t, mapVal)

		mapVal, err = FromString[map[string]any]("[]")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type map[string]interface {}: json: cannot unmarshal array into Go value of type map[string]interface {}")
		assert.Empty(t, mapVal)

		sliceVal, err := FromString[[]string]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type []string: invalid character 'i' looking for beginning of value")
		assert.Nil(t, sliceVal)

		sliceVal, err = FromString[[]string]("{}")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type []string: json: cannot unmarshal object into Go value of type []string")
		assert.Nil(t, sliceVal)

		structVal, err := FromString[struct{}]("invalid")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type struct {}: invalid character 'i' looking for beginning of value")
		assert.Empty(t, structVal)

		structVal, err = FromString[struct{}]("[]")
		require.Error(t, err)
		require.EqualError(t, err, "failed to convert string to type struct {}: json: cannot unmarshal array into Go value of type struct {}")
		assert.Empty(t, structVal)
	})

	t.Run("Valid", func(t *testing.T) {
		stringVal, err := FromString[string]("abcd")
		require.NoError(t, err)
		assert.Equal(t, "abcd", stringVal)

		intVal, err := FromString[int]("1234")
		require.NoError(t, err)
		assert.Equal(t, 1234, intVal)

		int8Val, err := FromString[int8]("123")
		require.NoError(t, err)
		assert.Equal(t, int8(123), int8Val)

		int16Val, err := FromString[int16]("1234")
		require.NoError(t, err)
		assert.Equal(t, int16(1234), int16Val)

		int32Val, err := FromString[int32]("1234")
		require.NoError(t, err)
		assert.Equal(t, int32(1234), int32Val)

		int64Val, err := FromString[int64]("1234")
		require.NoError(t, err)
		assert.Equal(t, int64(1234), int64Val)

		uintVal, err := FromString[uint]("1234")
		require.NoError(t, err)
		assert.Equal(t, uint(1234), uintVal)

		uint8Val, err := FromString[uint8]("123")
		require.NoError(t, err)
		assert.Equal(t, uint8(123), uint8Val)

		uint16Val, err := FromString[uint16]("1234")
		require.NoError(t, err)
		assert.Equal(t, uint16(1234), uint16Val)

		uint32Val, err := FromString[uint32]("1234")
		require.NoError(t, err)
		assert.Equal(t, uint32(1234), uint32Val)

		uint64Val, err := FromString[uint64]("1234")
		require.NoError(t, err)
		assert.Equal(t, uint64(1234), uint64Val)

		float32Val, err := FromString[float32]("1234.56")
		require.NoError(t, err)
		assert.InDelta(t, float32(1234.56), float32Val, 0.0001)

		float64Val, err := FromString[float64]("1234.56")
		require.NoError(t, err)
		assert.InDelta(t, float64(1234.56), float64Val, 0.0001)

		complex64Val, err := FromString[complex64]("1234.56")
		require.NoError(t, err)
		assert.Equal(t, complex64(1234.56), complex64Val)

		complex128Val, err := FromString[complex128]("1234.56")
		require.NoError(t, err)
		assert.Equal(t, complex128(1234.56), complex128Val)

		boolVal, err := FromString[bool]("true")
		require.NoError(t, err)
		assert.True(t, boolVal)

		bytesVal, err := FromString[[]byte]("abcd")
		require.NoError(t, err)
		assert.Equal(t, []byte("abcd"), bytesVal)

		runesVal, err := FromString[[]rune]("abcd")
		require.NoError(t, err)
		assert.Equal(t, []rune("abcd"), runesVal)

		builderVal, err := FromString[*strings.Builder]("abcd")
		require.NoError(t, err)
		assert.Equal(t, "abcd", builderVal.String())

		arrayVal, err := FromString[[]int]("[1,2,3,4]")
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4}, arrayVal)

		mapVal, err := FromString[map[string]int](`{"a":1,"b":2,"c":3,"d":4}`)
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}, mapVal)

		structVal, err := FromString[struct{ A int }](`{"A":1}`)
		require.NoError(t, err)
		assert.Equal(t, struct{ A int }{A: 1}, structVal)
	})
}

func TestFromStringOrEmpty(t *testing.T) {
	t.Run("Unsupported", func(t *testing.T) {
		assert.Nil(t, FromStringOrEmpty[func()](""))
		assert.Nil(t, FromStringOrEmpty[map[string]any](""))
		assert.Empty(t, FromStringOrEmpty[map[string]any](""))
		assert.Nil(t, FromStringOrEmpty[[]string](""))
		assert.Empty(t, FromStringOrEmpty[[]string](""))
		assert.Empty(t, FromStringOrEmpty[struct{}](""))
	})

	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, FromStringOrEmpty[func()]("abcd"))
		assert.Nil(t, FromStringOrEmpty[map[string]any]("abcd"))
		assert.Empty(t, FromStringOrEmpty[map[string]any]("abcd"))
		assert.Nil(t, FromStringOrEmpty[[]string]("abcd"))
		assert.Empty(t, FromStringOrEmpty[[]string]("abcd"))
		assert.Empty(t, FromStringOrEmpty[struct{}]("abcd"))
		assert.Equal(t, "", FromStringOrEmpty[string](""))
		assert.Zero(t, FromStringOrEmpty[int](""))
		assert.Zero(t, FromStringOrEmpty[int8](""))
		assert.Zero(t, FromStringOrEmpty[int16](""))
		assert.Zero(t, FromStringOrEmpty[int32](""))
		assert.Zero(t, FromStringOrEmpty[int64](""))
		assert.Zero(t, FromStringOrEmpty[uint](""))
		assert.Zero(t, FromStringOrEmpty[uint8](""))
		assert.Zero(t, FromStringOrEmpty[uint16](""))
		assert.Zero(t, FromStringOrEmpty[uint32](""))
		assert.Zero(t, FromStringOrEmpty[uint64](""))
		assert.Zero(t, FromStringOrEmpty[float32](""))
		assert.Zero(t, FromStringOrEmpty[float64](""))
		assert.Zero(t, FromStringOrEmpty[complex64](""))
		assert.Zero(t, FromStringOrEmpty[complex128](""))
		assert.False(t, FromStringOrEmpty[bool](""))
		assert.Empty(t, FromStringOrEmpty[[]byte](""))
		assert.Empty(t, FromStringOrEmpty[[]rune](""))
		assert.Equal(t, "", FromStringOrEmpty[*strings.Builder]("").String())
	})

	t.Run("Invalid", func(t *testing.T) {
		assert.Zero(t, FromStringOrEmpty[int]("invalid"))
		assert.Zero(t, FromStringOrEmpty[int8]("invalid"))
		assert.Zero(t, FromStringOrEmpty[int16]("invalid"))
		assert.Zero(t, FromStringOrEmpty[int32]("invalid"))
		assert.Zero(t, FromStringOrEmpty[int64]("invalid"))
		assert.Zero(t, FromStringOrEmpty[uint]("invalid"))
		assert.Zero(t, FromStringOrEmpty[uint8]("invalid"))
		assert.Zero(t, FromStringOrEmpty[uint16]("invalid"))
		assert.Zero(t, FromStringOrEmpty[uint32]("invalid"))
		assert.Zero(t, FromStringOrEmpty[uint64]("invalid"))
		assert.Zero(t, FromStringOrEmpty[float32]("invalid"))
		assert.Zero(t, FromStringOrEmpty[float64]("invalid"))
		assert.Zero(t, FromStringOrEmpty[complex64]("invalid"))
		assert.Zero(t, FromStringOrEmpty[complex128]("invalid"))
		assert.False(t, FromStringOrEmpty[bool]("invalid"))
		assert.Empty(t, FromStringOrEmpty[map[string]any]("invalid"))
		assert.Empty(t, FromStringOrEmpty[map[string]any]("[]"))
		assert.Empty(t, FromStringOrEmpty[[]string]("invalid"))
		assert.Empty(t, FromStringOrEmpty[[]string]("{}"))
		assert.Empty(t, FromStringOrEmpty[struct{}]("invalid"))
		assert.Empty(t, FromStringOrEmpty[struct{}]("[]"))
	})

	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "abcd", FromStringOrEmpty[string]("abcd"))
		assert.Equal(t, 1234, FromStringOrEmpty[int]("1234"))
		assert.Equal(t, int8(123), FromStringOrEmpty[int8]("123"))
		assert.Equal(t, int16(1234), FromStringOrEmpty[int16]("1234"))
		assert.Equal(t, int32(1234), FromStringOrEmpty[int32]("1234"))
		assert.Equal(t, int64(1234), FromStringOrEmpty[int64]("1234"))
		assert.Equal(t, uint(1234), FromStringOrEmpty[uint]("1234"))
		assert.Equal(t, uint8(123), FromStringOrEmpty[uint8]("123"))
		assert.Equal(t, uint16(1234), FromStringOrEmpty[uint16]("1234"))
		assert.Equal(t, uint32(1234), FromStringOrEmpty[uint32]("1234"))
		assert.Equal(t, uint64(1234), FromStringOrEmpty[uint64]("1234"))
		assert.InDelta(t, float32(1234.56), FromStringOrEmpty[float32]("1234.56"), 0.0001)
		assert.InDelta(t, float64(1234.56), FromStringOrEmpty[float64]("1234.56"), 0.0001)
		assert.Equal(t, complex64(1234.56), FromStringOrEmpty[complex64]("1234.56"))
		assert.Equal(t, complex128(1234.56), FromStringOrEmpty[complex128]("1234.56"))
		assert.True(t, FromStringOrEmpty[bool]("true"))
		assert.Equal(t, []byte("abcd"), FromStringOrEmpty[[]byte]("abcd"))
		assert.Equal(t, []rune("abcd"), FromStringOrEmpty[[]rune]("abcd"))
		assert.Equal(t, "abcd", FromStringOrEmpty[*strings.Builder]("abcd").String())
	})
}
