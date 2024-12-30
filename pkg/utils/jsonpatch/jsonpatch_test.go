package jsonpatch

import (
	"encoding/json"
	"testing"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestJSONPatchReplace(t *testing.T) {
	patch, err := jsonpatch.DecodePatch(NewPatches(
		NewReplace("/model", "gpt-3.5-turbo"),
	))
	require.NoError(t, err)

	patched, err := patch.Apply(lo.Must(json.Marshal(map[string]interface{}{
		"model": "gpt-3.5",
	})))
	require.NoError(t, err)

	require.JSONEq(t, `{"model":"gpt-3.5-turbo"}`, string(patched))
}

func TestJSONPatchAdd(t *testing.T) {
	patch, err := jsonpatch.DecodePatch(NewPatches(
		NewAdd("/stream_options", map[string]any{
			"include_usage": true,
		}),
	))
	require.NoError(t, err)

	patched, err := patch.Apply(lo.Must(json.Marshal(map[string]interface{}{
		"model": "gpt-3.5",
	})))
	require.NoError(t, err)

	require.JSONEq(t, `{"model":"gpt-3.5","stream_options":{"include_usage":true}}`, string(patched))
}

func TestJSONPatchRemove(t *testing.T) {
	patch, err := jsonpatch.DecodePatch(NewPatches(
		NewRemove("/model"),
	))
	require.NoError(t, err)

	patched, err := patch.Apply(lo.Must(json.Marshal(map[string]interface{}{
		"model": "gpt-3.5",
	})))
	require.NoError(t, err)

	require.JSONEq(t, `{}`, string(patched))
}
