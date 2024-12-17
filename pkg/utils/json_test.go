package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONPathExecute(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name     string
			payload  map[string]any
			template string
			expected any
		}

		testCases := []testCase{
			{
				name: "model",
				payload: map[string]any{
					"model": "gpt-4o",
				},
				template: "{ .model }",
				expected: "gpt-4o",
			},
			{
				name: "message role",
				payload: map[string]any{
					"model": "gpt-4o",
					"messages": []any{
						map[string]any{
							"role":    "user",
							"content": "Hello",
						},
					},
				},
				template: "{ .messages[0].role }",
				expected: "user",
			},
			{
				name: "message content",
				payload: map[string]any{
					"model": "gpt-4o",
					"messages": []any{
						map[string]any{
							"role":    "user",
							"content": "Hello",
						},
					},
				},
				template: "{ .messages[0].content }",
				expected: "Hello",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.expected, GetByJSONPath[string](tc.payload, tc.template))
			})
		}
	})

	t.Run("number", func(t *testing.T) {
		t.Parallel()

		payload := map[string]any{
			"code": 401,
		}

		assert.Equal(t, 401, GetByJSONPath[int](payload, "{ .code }"))
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()

		t.Run("unknown nil", func(t *testing.T) {
			t.Parallel()

			payload := map[string]any{
				"code": nil,
			}

			assert.Equal(t, "", GetByJSONPath[string](payload, "{ .code }"))
		})

		t.Run("nil string", func(t *testing.T) {
			t.Parallel()

			type payload struct {
				Code *string `json:"code"`
			}

			p := payload{
				Code: nil,
			}

			assert.Equal(t, "", GetByJSONPath[string](p, "{ .code }"))
		})
	})
}
